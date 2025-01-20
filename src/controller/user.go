package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
)

func userGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/user")
	errWarper := dtoError.GetServiceErrorWarpper()
	group.POST("/register", func(c *gin.Context) {
		var req dto.UserRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		_, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			serviceErr := errWarper.NewParseFormatFailedServiceError(err, "Birthday should be YYYY-MM-DD")
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		res, serviceErr := service.GetAccountService().UserRegisterService(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.POST("/login", func(c *gin.Context) {
		var req dto.UserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		res, serviceErr := service.GetAccountService().UserLoginService(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		SetSessionValue(c, res.ID, res.Username)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": dtoError.Success,
		})
	})

	group.PUT("/reset_password", func(c *gin.Context) {
		var req dto.ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		serviceErr := service.GetAccountService().ResetPasswordService(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errorCode": dtoError.Success,
		})
	})

	group.Use(NewLoginFilter())

	group.GET("/user_info", func(c *gin.Context) {
		req := dto.GetUserInfoRequest{}
		_, id, _ := GetSessionValue(c)
		req.ID = id
		res, serviceErr := service.GetAccountService().UserInfoService(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": res})
	})
}
