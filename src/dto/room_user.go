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

type FetchInvitationByAdminRequest struct {
	AdminID uint64
	RoomID  uint64 `json:"room_id" binding:"required"`
}

type FetchInvitationByAdminResponse struct {
	RoomID    uint64     `json:"room_id" binding:"required"`
	UserInfos []UserInfo `json:"uesr_info" binding:"required"`
}

type FetchInvitationByUserRequest struct {
	UserID uint64
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
}

type FetchApplicationByAdminResponse struct {
	RoomID    uint64     `json:"room_id" binding:"required"`
	UserInfos []UserInfo `json:"user_infos" binding:"required"`
}

type FetchApplicationByUserRequest struct {
	UserID uint64
}

type FetchApplicationByUserResponse struct {
	UserID    uint64     `json:"user_id" binding:"required"`
	RoomInfos []RoomInfo `json:"room_infos" binding:"required"`
}

type DeleteUserRequest struct {
	RoomID      uint64 `json:"room_id" binding:"required"`
	AdminUserID uint64
	UserID      uint64 `json:"user_id" binding:"required"`
}

type DeleteUserResponse struct {
}
