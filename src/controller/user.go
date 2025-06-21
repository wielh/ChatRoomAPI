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
	group.POST("/register", user.Register)
	group.POST("/login", user.Login)
	group.PUT("/reset_password", user.ResetPassword)
	group.Use(GetLoginFilter())
	group.GET("/info", user.GetUserInfo)
}

type UserController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	ResetPassword(c *gin.Context)
	GetUserInfo(c *gin.Context)
}

type UserControllerImpl struct {
	errWarper dtoError.ServiceErrorWarpper
}

var user UserController

func init() {
	user = &UserControllerImpl{
		errWarper: dtoError.GetServiceErrorWarpper(),
	}
}

func (u *UserControllerImpl) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := u.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		serviceErr := u.errWarper.NewParseFormatFailedServiceError(err, "Birthday should be YYYY-MM-DD")
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	res, serviceErr := service.GetAccountService().UserRegisterService(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (u *UserControllerImpl) Login(c *gin.Context) {
	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := u.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	res, serviceErr := service.GetAccountService().UserLoginService(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, err := SetSessionValue(c, res.ID, res.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (u *UserControllerImpl) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := u.errWarper.NewParseJsonFailedServiceError(err)
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
}

func (u *UserControllerImpl) GetUserInfo(c *gin.Context) {
	req := dto.GetUserInfoRequest{}
	_, id, _ := GetSessionValue(c)
	req.ID = id
	res, serviceErr := service.GetAccountService().UserInfoService(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": res})
}
