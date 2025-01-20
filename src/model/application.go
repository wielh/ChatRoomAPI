package model

import "github.com/lib/pq"

type ApplyRecord struct {
	Id     uint64 `gorm:"primaryKey;column:id"`
	RoomID uint64 `gorm:"not null;column:room_id"`
	UserID uint64 `gorm:"not null;column:user_id"`
}

type RoomJoinApplyRecord struct {
	Id          uint64
	RoomId      uint64
	RoomName    string
	AdminUserId uint64
	UserIds     pq.Int64Array `gorm:"type:bigint[]"`
	Description string
}

type UserJoinApplyRecord struct {
	Id     uint64
	UserId uint64
	Name   string
	Email  string
}
