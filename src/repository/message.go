package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type MessageRepository interface {
	AddMessage(ctx context.Context, roomID uint64, userID uint64, content string) (*model.Message, error)
	FetchMessages(ctx context.Context, roomID uint64, TimeCursor time.Time,
		resultMaxSize int32) (messages []*model.Message, NextTimeCursor time.Time, err error)
}

type messageRepositoryImpl struct {
	DB *gorm.DB
}

var message MessageRepository

func init() {
	message = &messageRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetMessageRepository() MessageRepository {
	return message
}

func (m *messageRepositoryImpl) AddMessage(ctx context.Context, roomID uint64, userID uint64, content string) (*model.Message, error) {
	tx := GetTxContext(ctx, m.DB)
	message := model.Message{RoomID: roomID, UserID: userID, Content: content}
	result := tx.Create(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &message, nil
}

func (m *messageRepositoryImpl) FetchMessages(ctx context.Context, roomID uint64, TimeCursor time.Time,
	resultMaxSize int32) (messages []*model.Message, NextTimeCursor time.Time, err error) {

	if resultMaxSize == 0 {
		return messages, TimeCursor, nil
	}

	tx := GetTxContext(ctx, m.DB)
	columns := []string{"id", "room_id", "user_id", "content", "create_time"}
	var result *gorm.DB
	if resultMaxSize > 0 {
		result = tx.Select(columns).Where("create_time > ?", TimeCursor).Order("create_time ASC").Limit(int(resultMaxSize)).Find(&messages)
	} else {
		result = tx.Select(columns).Where("create_time < ?", TimeCursor).Order("create_time DESC").Limit(int(-1 * resultMaxSize)).Find(&messages)
	}

	if result.Error != nil {
		err = result.Error
		return
	} else if len(messages) == 0 {
		NextTimeCursor = TimeCursor
		return
	}
	NextTimeCursor = messages[len(messages)-1].CreatedAt
	return
}
