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

func messageGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/message")
	errWarper := dtoError.GetServiceErrorWarpper()
	group.Use(NewLoginFilter())

	group.POST("/", func(c *gin.Context) {
		var req dto.AddMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.GET("/", func(c *gin.Context) {
		req := dto.FetchMessageRequest{MessageSize: 100, TimeCursor: common.TimeToUint64(time.Now())} //default value
		if err := c.ShouldBindQuery(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})
}
