package controller

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func roomGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/room")
	group.Use(GetLoginFilter())

	group.PUT("/", room.CreateRoom)
	group.GET("/", room.GetAvailbleRooms)
	group.GET("/info", room.GetRoomInfo)
	group.DELETE("/", room.DeleteRoom)

	roomAdminGroupRouter(group)
	roomUserGroupRouter(group)
}

type RoomController interface {
	CreateRoom(c *gin.Context)
	GetAvailbleRooms(c *gin.Context)
	GetRoomInfo(c *gin.Context)
	DeleteRoom(c *gin.Context)
}

type roomControllerImpl struct {
	errWarper   dtoError.ServiceErrorWarpper
	roomService service.RoomService
}

var room RoomController

func init() {
	room = &roomControllerImpl{
		errWarper:   dtoError.GetServiceErrorWarpper(),
		roomService: service.GetRoomService(),
	}
}

func (r *roomControllerImpl) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	res, serviceErr := r.roomService.CreateRoom(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomControllerImpl) GetAvailbleRooms(c *gin.Context) {
	_, userId, _ := GetSessionValue(c)
	req := dto.GetAvailbleRoomsRequest{UserID: userId}
	res, serviceErr := r.roomService.GetAvailbleRooms(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomControllerImpl) GetRoomInfo(c *gin.Context) {
	_, userId, _ := GetSessionValue(c)
	roomIDStr := c.Query("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		serviceErr := r.errWarper.NewParseFormatFailedServiceError(err, "invaild room_id")
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	req := dto.ReadRoomInfoRequest{UserID: userId, RoomID: roomID}
	res, serviceErr := r.roomService.ReadRoomInfo(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomControllerImpl) DeleteRoom(c *gin.Context) {
	var req dto.DeleteRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId
	_, serviceErr := r.roomService.DeleteRoom(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
