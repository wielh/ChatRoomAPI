package controller

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/logger"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func commonMiddleware(g *gin.RouterGroup) {
	GetLoginFilter()
	rateLimit := src.GlobalConfig.YamlConfig.Server.RateLimitConfig
	coustomAllLimiter := newRateLimiter(
		rateLimit.All.MaxRequest, rateLimit.All.Second, func(c *gin.Context) string {
			return "<<<ALL>>>"
		},
	)
	coustomIPLimiter := newRateLimiter(
		rateLimit.IP.MaxRequest, rateLimit.IP.Second, func(c *gin.Context) string {
			return c.ClientIP()
		},
	)
	customRepeatedLimiter := newRateLimiter(
		rateLimit.Repeat.MaxRequest, rateLimit.Repeat.Second, func(c *gin.Context) string {
			return fmt.Sprintf("%s::%s", c.ClientIP(), c.Request.URL)
		},
	)
	g.Use(
		customRepeatedLimiter,
		coustomIPLimiter,
		coustomAllLimiter,
		customTimeout(),
		customRequestUUIDGenerator(),
		customLogger(),
		customRecovery(),
		readLoginSession,
	)
}

// ==============================================================================================
// ===
var readLoginSession gin.HandlerFunc
var loginFilter func(*gin.Context)
var log = logger.NewLogger()
var once sync.Once

func GetLoginFilter() func(*gin.Context) {
	once.Do(func() {
		readLoginSession = sessions.Sessions("login", *src.GlobalConfig.RedisSession)
		loginFilter = func(c *gin.Context) {
			ok, _, _ := GetSessionValue(c)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
				c.Abort()
				return
			}
			c.Next()
		}
	})
	return loginFilter
}

func SetSessionValue(c *gin.Context, ID uint64, username string) {
	session := sessions.Default(c)
	session.Set("id", ID)
	session.Set("username", username)
	session.Save()
}

func GetSessionValue(c *gin.Context) (bool, uint64, string) {
	session := sessions.Default(c)
	id, ok1 := session.Get("id").(uint64)
	username, ok2 := session.Get("username").(string)
	return ok1 && ok2, id, username
}

// ===
// ==============================================================================================

func customRequestUUIDGenerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		common.SetUUID(c)
		c.Next()
	}
}

func newRateLimiter(maxRequests int, windowSeconds int, keyGenerator func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		redisKey := keyGenerator(c)
		rdb := src.GlobalConfig.Redis

		currentCount, err := rdb.Incr(c, redisKey).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if currentCount == 1 {
			err = rdb.Expire(c, redisKey, time.Duration(windowSeconds)*time.Second).Err()
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
		}

		if int(currentCount) > maxRequests {
			c.AbortWithStatus(429)
			return
		}
		c.Next()
	}
}

func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		data := map[string]any{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"clientIP": c.ClientIP(),
			"start":    start,
			"duration": duration,
			"errors":   c.Errors,
		}

		if len(c.Errors) > 0 {
			log.Error(common.GetUUID(c), "customRecoveryError", data, nil)
		} else {
			log.Info(common.GetUUID(c), "customLoggerInfo", data, nil)
		}
	}
}

func customRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var panicMessage string
				switch v := err.(type) {
				case string:
					panicMessage = v
				case error:
					panicMessage = v.Error()
				default:
					panicMessage = fmt.Sprintf("Unknown panic: %v", v)
				}

				data := map[string]any{
					"method":   c.Request.Method,
					"path":     c.Request.URL.Path,
					"clientIP": c.ClientIP(),
					"panic":    panicMessage,
				}
				log.Error(common.GetUUID(c), "customRecovery", data, nil)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

func customTimeout() gin.HandlerFunc {
	timeoutSecond := time.Duration(src.GlobalConfig.YamlConfig.Server.Timeout)
	return timeout.New(
		timeout.WithTimeout(timeoutSecond*time.Second),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(func(c *gin.Context) {
			c.String(http.StatusRequestTimeout, "timeout")
		}),
	)
}
