package model

import "github.com/lib/pq"

type InviteRecord struct {
	Id     uint64 `gorm:"primaryKey;column:id"`
	RoomID uint64 `gorm:"not null;column:room_id"`
	UserID uint64 `gorm:"not null;column:user_id"`
}

type RoomJoinInviteRecord struct {
	Id          uint64
	RoomId      uint64
	RoomName    string
	AdminUserId uint64
	UserIds     pq.Int64Array `gorm:"type:bigint[]"`
	Description string
}

type UserJoinInviteRecord struct {
	Id     uint64
	UserId uint64
	Name   string
	Email  string
}
