package model

type Message struct {
	ID      uint64 `gorm:"primaryKey;column:id"`
	RoomID  uint64 `gorm:"not null;column:room_id"`
	UserID  uint64 `gorm:"not null;column:user_id"`
	Content string `gorm:"not null;column:content"`
	Base
}
