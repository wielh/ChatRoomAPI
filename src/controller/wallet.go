package controller

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WalletRouter(g *gin.RouterGroup) {
	group := g.Group("/wallet")
	group.Use(GetLoginFilter())
	group.GET("/", wallet.GetState)
	group.POST("/", wallet.Charge)
}

type WalletController interface {
	GetState(c *gin.Context)
	Charge(c *gin.Context)
}

type walletControllerImpl struct {
	errWarper     dtoError.ServiceErrorWarpper
	walletService service.WalletService
}

func (w *walletControllerImpl) Charge(c *gin.Context) {
	var req dto.ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := w.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	res, serviceErr := w.walletService.Charge(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (w *walletControllerImpl) GetState(c *gin.Context) {
	var req dto.GetStateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := w.errWarper.NewParseQueryFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	res, serviceErr := w.walletService.GetState(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

var wallet WalletController

func init() {
	wallet = &walletControllerImpl{
		errWarper:     dtoError.GetServiceErrorWarpper(),
		walletService: service.GetWalletService(),
	}
}
