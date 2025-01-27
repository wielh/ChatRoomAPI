package controller

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func roomUserGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/user")

	group.PUT("/invitation_confrim", roomUser.InvitationConfrim)
	group.GET("/invitations", roomUser.InvitationConfrim)
	group.PUT("/application", roomUser.Apply)
	group.DELETE("/application", roomUser.DeleteApplication)
	group.GET("/applications", roomUser.FetchApplications)
}

type RoomUserController interface {
	InvitationConfrim(c *gin.Context)
	FetchInvitations(c *gin.Context)
	Apply(c *gin.Context)
	DeleteApplication(c *gin.Context)
	FetchApplications(c *gin.Context)
}

type roomUserControllerImpl struct {
	errWarpper dtoError.ServiceErrorWarpper
}

var roomUser RoomUserController

func init() {
	roomUser = &roomUserControllerImpl{
		errWarpper: dtoError.GetServiceErrorWarpper(),
	}
}

func (r *roomUserControllerImpl) InvitationConfrim(c *gin.Context) {
	var req dto.ConfrimInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	_, serviceErr := service.GetRoomUserService().ConfrimInvite(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

func (r *roomUserControllerImpl) FetchInvitations(c *gin.Context) {
	req := dto.FetchInvitationByUserRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId

	res, serviceErr := service.GetRoomUserService().FetchInvitationsByUser(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomUserControllerImpl) Apply(c *gin.Context) {
	var req dto.RoomJoinApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	_, serviceErr := service.GetRoomUserService().RoomJoinApply(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})

}

func (r *roomUserControllerImpl) DeleteApplication(c *gin.Context) {
	var req dto.RoomJoinApplyCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": 0,
		})
	}
	_, userId, _ := GetSessionValue(c)
	req.UserID = userId

	_, serviceErr := service.GetRoomUserService().RoomJoinApplyCancel(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

func (r *roomUserControllerImpl) FetchApplications(c *gin.Context) {
	req := dto.FetchApplicationByUserRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.UserID = userId
	res, serviceErr := service.GetRoomUserService().FetchApplicationByUser(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}
