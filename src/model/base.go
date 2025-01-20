package model

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	CreatedAt time.Time      `gorm:"not null;autoCreateTime:nano;column:create_time"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime:nano;column:update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:delete_time"`
}
