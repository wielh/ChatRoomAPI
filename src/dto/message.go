package dto

type AddMessageRequest struct {
	RoomID  uint64 `json:"room_id" binding:"required"`
	UserID  uint64
	Content string `json:"content" binding:"required"`
}

type AddMessageResponse struct {
	ID        uint64 `json:"id" binding:"required"`
	CreatedAt uint64 `json:"create_time" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

type FetchMessageRequest struct {
	RoomID      uint64 `form:"room_id" binding:"required"`
	UserID      uint64
	TimeCursor  uint64 `form:"time_cursor"`
	MessageSize int32  `form:"message_size"`
}

type Message struct {
	ID        uint64 `json:"id" binding:"required"`
	UserID    uint64 `json:"user_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
	CreatedAt uint64 `json:"create_time" binding:"required"`
}

type FetchMessageResponse struct {
	NextTimeCursor uint64    `json:"next_time_cursor" binding:"required"`
	Messages       []Message `json:"messages" binding:"required"`
}
