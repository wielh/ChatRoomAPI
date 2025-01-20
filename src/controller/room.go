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
	errWarper := dtoError.GetServiceErrorWarpper()
	group.Use(NewLoginFilter())

	group.PUT("/", func(c *gin.Context) {
		var req dto.CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		_, userId, _ := GetSessionValue(c)
		req.UserID = userId
		res, serviceErr := service.GetRoomService().CreateRoom(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.GET("/", func(c *gin.Context) {
		_, userId, _ := GetSessionValue(c)
		req := dto.GetAvailbleRoomsRequest{UserID: userId}
		res, serviceErr := service.GetRoomService().GetAvailbleRooms(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.GET("/info", func(c *gin.Context) {
		_, userId, _ := GetSessionValue(c)
		roomIDStr := c.Query("room_id")
		roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
		if err != nil {
			serviceErr := errWarper.NewParseFormatFailedServiceError(err, "invaild room_id")
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		req := dto.ReadRoomInfoRequest{UserID: userId, RoomID: roomID}
		res, serviceErr := service.GetRoomService().ReadRoomInfo(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.DELETE("/", func(c *gin.Context) {
		var req dto.DeleteRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		_, userId, _ := GetSessionValue(c)
		req.AdminUserID = userId
		_, serviceErr := service.GetRoomService().DeleteRoom(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		c.JSON(http.StatusNoContent, gin.H{})
	})

	roomAdminGroupRouter(group)
	roomUserGroupRouter(group)
}

func roomAdminGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/admin")
	errWarper := dtoError.GetServiceErrorWarpper()

	group.PATCH("/", func(c *gin.Context) {
		var req dto.AdminChangeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.PUT("/invitation", func(c *gin.Context) {
		var req dto.InviteNewUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.DELETE("/invitation", func(c *gin.Context) {
		var req dto.InviteNewUserCancelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.GET("/invitations", func(c *gin.Context) {
		_, userId, _ := GetSessionValue(c)
		req := dto.FetchInvitationByAdminRequest{AdminID: userId}
		res, serviceErr := service.GetRoomAdminService().FetchInvitationsByAdmin(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.PUT("/application_confrim", func(c *gin.Context) {
		var req dto.ConfrimApplyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.GET("/applications", func(c *gin.Context) {
		_, userId, _ := GetSessionValue(c)
		roomIDStr := c.Query("room_id")
		roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
		if err != nil {
			serviceErr := errWarper.NewParseFormatFailedServiceError(err, "invaild roomId")
			c.JSON(serviceErr.ToJsonResponse())
			return
		}

		req := dto.FetchApplicationByAdminRequest{AdminUserID: userId, RoomID: roomID}
		res, serviceErr := service.GetRoomAdminService().FetchApplicationByAdmin(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.DELETE("/user", func(c *gin.Context) {
		var req dto.DeleteUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})
}

func roomUserGroupRouter(g *gin.RouterGroup) {
	group := g.Group("/user")
	errWarper := dtoError.GetServiceErrorWarpper()

	group.PUT("/invitation_confrim", func(c *gin.Context) {
		var req dto.ConfrimInviteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.GET("/invitations", func(c *gin.Context) {
		req := dto.FetchInvitationByUserRequest{}
		_, userId, _ := GetSessionValue(c)
		req.UserID = userId

		res, serviceErr := service.GetRoomUserService().FetchInvitationsByUser(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})

	group.PUT("/application", func(c *gin.Context) {
		var req dto.RoomJoinApplyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			serviceErr := errWarper.NewParseJsonFailedServiceError(err)
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
	})

	group.DELETE("/application", func(c *gin.Context) {
		var req dto.RoomJoinApplyCancelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": 0,
			})
			return
		}

		_, userId, _ := GetSessionValue(c)
		req.UserID = userId
		_, serviceErr := service.GetRoomUserService().RoomJoinApplyCancel(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})
	})

	group.GET("/applications", func(c *gin.Context) {
		req := dto.FetchApplicationByUserRequest{}
		_, userId, _ := GetSessionValue(c)
		req.UserID = userId
		res, serviceErr := service.GetRoomUserService().FetchApplicationByUser(c, &req)
		if serviceErr != nil {
			c.JSON(serviceErr.ToJsonResponse())
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": res})
	})
}
