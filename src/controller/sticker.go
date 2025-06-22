package controller

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func stickerRouter(g *gin.RouterGroup) {
	group := g.Group("/sticker")
	group.Use(GetLoginFilter())
	group.GET("/info", sticker.GetStickerSetInfo)
	group.GET("/list", sticker.GetAllAvailableStickersInfo)
	group.PUT("/buy", sticker.BuyStickerSet)
}

type StickerController interface {
	GetStickerSetInfo(c *gin.Context)
	BuyStickerSet(c *gin.Context)
	GetAllAvailableStickersInfo(c *gin.Context)
}

type stickerControllerImpl struct {
	errWarper      dtoError.ServiceErrorWarpper
	stickerService service.StickerService
}

var sticker StickerController

func init() {
	sticker = &stickerControllerImpl{
		errWarper:      dtoError.GetServiceErrorWarpper(),
		stickerService: service.GetStickerService(),
	}
}

func (s *stickerControllerImpl) BuyStickerSet(c *gin.Context) {
	var req dto.BuyStickerSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := s.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	res, serviceErr := s.stickerService.BuyStickerSet(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *stickerControllerImpl) GetStickerSetInfo(c *gin.Context) {
	var req dto.GetStickerSetInfoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := s.errWarper.NewParseQueryFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	res, serviceErr := s.stickerService.GetStickerSetInfo(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (s *stickerControllerImpl) GetAllAvailableStickersInfo(c *gin.Context) {
	_, userId, _ := GetSessionValue(c)
	req := dto.GetAllAvailableStickersInfoRequest{UserID: userId}
	res, serviceErr := s.stickerService.GetAllAvailableStickersInfo(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}
