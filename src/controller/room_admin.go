package controller

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func roomAdminGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/admin")

	group.PATCH("/", roomAdmin.AdminChange)
	group.PUT("/invitation", roomAdmin.Invite)
	group.DELETE("/invitation", roomAdmin.DeleteInvitation)
	group.GET("/invitations", roomAdmin.FetchInvitations)
	group.PUT("/application_confrim", roomAdmin.ConfrimApplication)
	group.GET("/applications", roomAdmin.FetchApplications)
	group.DELETE("/user", roomAdmin.DeleteUser)
}

type RoomAdminController interface {
	AdminChange(c *gin.Context)
	Invite(c *gin.Context)
	DeleteInvitation(c *gin.Context)
	FetchInvitations(c *gin.Context)
	ConfrimApplication(c *gin.Context)
	FetchApplications(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type roomAdminControllerImpl struct {
	errWarpper dtoError.ServiceErrorWarpper
}

var roomAdmin RoomAdminController

func init() {
	roomAdmin = &roomAdminControllerImpl{
		errWarpper: dtoError.GetServiceErrorWarpper(),
	}
}

func (r *roomAdminControllerImpl) AdminChange(c *gin.Context) {
	var req dto.AdminChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId
	res, serviceErr := service.GetRoomAdminService().AdminChange(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomAdminControllerImpl) Invite(c *gin.Context) {
	var req dto.InviteNewUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId

	_, serviceErr := service.GetRoomAdminService().InviteNewUser(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (r *roomAdminControllerImpl) DeleteInvitation(c *gin.Context) {
	var req dto.InviteNewUserCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId
	_, serviceErr := service.GetRoomAdminService().InviteNewUserCancel(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (r *roomAdminControllerImpl) FetchInvitations(c *gin.Context) {
	req := dto.FetchInvitationByAdminRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := r.errWarpper.NewParseQueryFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminID = userId

	res, serviceErr := service.GetRoomAdminService().FetchInvitationsByAdmin(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomAdminControllerImpl) ConfrimApplication(c *gin.Context) {
	var req dto.ConfrimApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId
	_, serviceErr := service.GetRoomAdminService().ConfrimApply(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (r *roomAdminControllerImpl) FetchApplications(c *gin.Context) {
	req := dto.FetchApplicationByAdminRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		serviceErr := r.errWarpper.NewParseQueryFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId

	res, serviceErr := service.GetRoomAdminService().FetchApplicationByAdmin(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": res})
}

func (r *roomAdminControllerImpl) DeleteUser(c *gin.Context) {
	var req dto.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		serviceErr := r.errWarpper.NewParseJsonFailedServiceError(err)
		c.JSON(serviceErr.ToJsonResponse())
		return
	}

	_, userId, _ := GetSessionValue(c)
	req.AdminUserID = userId
	_, serviceErr := service.GetRoomAdminService().DeleteUser(c, &req)
	if serviceErr != nil {
		c.JSON(serviceErr.ToJsonResponse())
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}
