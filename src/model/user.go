package model

import "time"

type User struct {
	Id       uint64    `gorm:"primaryKey;column:id"`
	Username string    `gorm:"not null;column:username"`
	Password string    `gorm:"not null;column:password"`
	Name     string    `gorm:"not null;column:name"`
	Birthday time.Time `gorm:"not null;column:birthday"`
	Email    string    `gorm:"not null;column:email"`
	Base
}
