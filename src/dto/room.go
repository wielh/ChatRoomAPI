package dto

type CreateRoomRequest struct {
	UserID      uint64
	RoomName    string `json:"room_name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreateRoomResponse struct {
	RoomID uint64 `json:"room_id" binding:"required"`
}

type GetAvailbleRoomsRequest struct {
	UserID   uint64
	Page     uint32 `form:"page" binding:"required,gte=1"`
	PageSize uint32 `form:"page_size" binding:"required,gte=1"`
}

type GetAvailbleRoomsResponse struct {
	RoomsInfos []ReadRoomInfoResponse `json:"room_infos" binding:"required"`
}

type ReadRoomInfoRequest struct {
	RoomID uint64 `json:"room_id" binding:"required"`
	UserID uint64
}

type ReadRoomInfoResponse struct {
	ID          uint64   `json:"room_id" binding:"required"`
	Name        string   `json:"room_name" binding:"required"`
	AdminUserID uint64   `json:"admin_user_id" binding:"required"`
	UserIDs     []uint64 `json:"userids" binding:"required"`
	Description string   `json:"description" binding:"required"`
}

type DeleteRoomRequest struct {
	AdminUserID uint64
	RoomID      uint64 `json:"room_id" binding:"required"`
}

type DeleteRoomResponse struct{}
