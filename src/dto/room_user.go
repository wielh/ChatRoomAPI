package dto

type ConfrimInviteRequest struct {
	RoomID  uint64 `json:"room_id" binding:"required"`
	UserID  uint64
	Allowed bool `json:"allow" binding:"required"`
}

type ConfrimInviteResponse struct{}

type UserInfo struct {
	UserID   uint64 `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type RoomInfo struct {
	RoomID      uint64   `json:"room_id" binding:"required"`
	RoomName    string   `json:"room_name" binding:"required"`
	AdminID     uint64   `json:"admin_user_id" binding:"required"`
	UserIDs     []uint64 `json:"user_ids" binding:"required"`
	Description string   `json:"description" binding:"required"`
}

type FetchInvitationByUserRequest struct {
	UserID   uint64
	Page     uint32 `form:"page" binding:"required,gte=1"`
	PageSize uint32 `form:"page_size" binding:"required,gte=1"`
}

type FetchInvitationByUserResponse struct {
	UserID    uint64     `json:"user_id" binding:"required"`
	RoomInfos []RoomInfo `json:"room_infos" binding:"required"`
}

type RoomJoinApplyRequest struct {
	RoomID uint64 `json:"room_id" binding:"required"`
	UserID uint64
}

type RoomJoinApplyResponse struct{}

type RoomJoinApplyCancelRequest struct {
	RoomID uint64 `json:"room_id" binding:"required"`
	UserID uint64
}

type RoomJoinApplyCancelResponse struct{}

type FetchApplicationByUserRequest struct {
	UserID   uint64
	Page     uint32 `form:"page" binding:"required,gte=1"`
	PageSize uint32 `form:"page_size" binding:"required,gte=1"`
}

type FetchApplicationByUserResponse struct {
	UserID    uint64     `json:"user_id" binding:"required"`
	RoomInfos []RoomInfo `json:"room_infos" binding:"required"`
}
