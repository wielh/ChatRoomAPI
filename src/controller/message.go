package controller

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageGroupController interface {
	AddMessage(c *gin.Context)
	FetchMessages(c *gin.Context)
}

type messageGroupControllerImpl struct {
	pageSize  int32
	errWarper dtoError.ServiceErrorWarpper
}

var message MessageGroupController

func init() {
	message = &messageGroupControllerImpl{
		errWarper: dtoError.GetServiceErrorWarpper(),
	}
}

func (m *messageGroupControllerImpl) AddMessage(c *gin.Context) {
	var req dto.AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := m.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	_, userId, _ := GetSessionValue(c)
	req.UserID = userId

	res, serviceErr := service.GetMessageService().AddMessage(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (m *messageGroupControllerImpl) FetchMessages(c *gin.Context) {
	req := dto.FetchMessageRequest{
		MessageSize: m.pageSize,
		TimeCursor:  common.TimeToUint64(time.Now()),
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := m.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	_, userId, _ := GetSessionValue(c)
	req.UserID = userId

	res, serviceErr := service.GetMessageService().FetchMessages(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func messageGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/message")
	group.Use(NewLoginFilter())
	group.POST("/", message.AddMessage)
	group.GET("/", message.FetchMessages)
}
