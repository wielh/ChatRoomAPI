package dto

type AdminChangeRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
}

type AdminChangeResponse struct{}

type InviteNewUserRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
}

type InviteNewUserResponse struct{}

type InviteNewUserCancelRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
}

type InviteNewUserCancelResponse struct{}

type FetchInvitationByAdminRequest struct {
	AdminID  uint64
	RoomID   uint64 `form:"room_id" binding:"required"`
	Page     uint32 `form:"page" binding:"required,gte=1"`
	PageSize uint32 `form:"page_size" binding:"required,gte=1"`
}

type FetchInvitationByAdminResponse struct {
	RoomID    uint64     `json:"room_id" binding:"required"`
	UserInfos []UserInfo `json:"uesr_info" binding:"required"`
}

type ConfrimApplyRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
	Allowed     bool   `json:"allow" binding:"required"`
}

type ConfrimApplyResponse struct {
}

type FetchApplicationByAdminRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	Page        uint32 `form:"page" binding:"required,gte=1"`
	PageSize    uint32 `form:"page_size" binding:"required,gte=1"`
}

type FetchApplicationByAdminResponse struct {
	RoomID    uint64     `json:"room_id" binding:"required"`
	UserInfos []UserInfo `json:"user_infos" binding:"required"`
}

type DeleteUserRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
}

type DeleteUserResponse struct {
}
