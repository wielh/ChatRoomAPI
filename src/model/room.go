package model

import (
	"github.com/lib/pq"
)

type Room struct {
	Id          uint64        `gorm:"primaryKey;column:id"`
	AdminUserID uint64        `gorm:"not null;column:admin_user_id"`
	Name        string        `gorm:"not null;column:name"`
	UserIDs     pq.Int64Array `gorm:"not null;column:user_ids;type:bigint[]"`
	Description string        `gorm:"not null;column:description"`
	Base
}
