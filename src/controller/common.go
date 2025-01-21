package controller

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/common"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	limit "github.com/yangxikun/gin-limit-by-key"
	"golang.org/x/time/rate"
)

func commonMiddleware(g *gin.RouterGroup) {
	NewLoginFilter()
	g.Use(
		customRepeatedLimiter(),
		customIPLimiter(),
		customAllLimiter(),
		customRequestUUIDGenerator(),
		customLogger(),
		customRecovery(),
		customTimeout(),
		readLoginSession,
	)
}

// ==============================================================================================
// ===
var readLoginSession gin.HandlerFunc
var loginFilter func(*gin.Context)
var logger = log.New(gin.DefaultWriter, "", log.LstdFlags)
var once sync.Once

func NewLoginFilter() func(*gin.Context) {
	once.Do(func() {
		store := cookie.NewStore([]byte(src.GlobalConfig.YamlConfig.Server.SessionEncryptKey))
		store.Options(sessions.Options{
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		readLoginSession = sessions.Sessions("login", store)
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

func customAllLimiter() gin.HandlerFunc {
	return limit.NewRateLimiter(func(c *gin.Context) string {
		return "<<<ALL>>>"
	}, func(c *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(time.Second), 1000), time.Hour
	}, func(c *gin.Context) {
		c.AbortWithStatus(429)
	})
}

func customIPLimiter() gin.HandlerFunc {
	return limit.NewRateLimiter(func(c *gin.Context) string {
		return c.ClientIP()
	}, func(c *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(time.Second), 20), time.Hour
	}, func(c *gin.Context) {
		c.AbortWithStatus(429)
	})
}

func customRepeatedLimiter() gin.HandlerFunc {
	return limit.NewRateLimiter(func(c *gin.Context) string {
		return fmt.Sprintf("%s::%s", c.ClientIP(), c.Request.URL)
	}, func(c *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(time.Second), 2), time.Hour
	}, func(c *gin.Context) {
		c.AbortWithStatus(429)
	})
}

func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if len(c.Errors) > 0 {
			logMessage := "fatal: [%s] %s | %3d | %13v | %15s | %s  | <<< %+v >>> \n"
			logger.Fatalf(logMessage, time.Now().Format(time.RFC3339), method, status, duration, clientIP, path, c.Errors)
		} else {
			logMessage := "info: [%s] %s | %3d | %13v | %15s | %s\n"
			logger.Printf(logMessage, time.Now().Format(time.RFC3339), method, status, duration, clientIP, path)
		}
	}
}

func customRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic: %v\nStack Trace: %s\n", err, debug.Stack())
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

func customTimeout() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(5000*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(func(c *gin.Context) {
			c.String(http.StatusRequestTimeout, "timeout")
		}),
	)
}
