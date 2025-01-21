package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
)

type MessageService interface {
	AddMessage(ctx context.Context, req *dto.AddMessageRequest) (*dto.AddMessageResponse, *dtoError.ServiceError)
	FetchMessages(ctx context.Context, req *dto.FetchMessageRequest) (*dto.FetchMessageResponse, *dtoError.ServiceError)
}

type messageServiceImpl struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
	errWarpper  dtoError.ServiceErrorWarpper
	logger      logger.Logger
}

var message MessageService

func init() {
	message = &messageServiceImpl{
		messageRepo: repository.GetMessageRepository(),
		roomRepo:    repository.GetRoomRepository(),
		errWarpper:  dtoError.GetServiceErrorWarpper(),
		logger:      logger.NewLogger(),
	}
}

func GetMessageService() MessageService {
	return message
}

func (m *messageServiceImpl) AddMessage(ctx context.Context, req *dto.AddMessageRequest) (*dto.AddMessageResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	m.logger.Info(requestId, "start", req, nil)
	defer func() { m.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	InRoom, err := m.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		m.logger.Error(requestId, "m.roomRepo.CheckUserInRoom", req, err)
		return nil, m.errWarpper.NewDBServiceError(err)
	} else if !InRoom {
		tx.Rollback()
		return nil, m.errWarpper.NewUserNotInRoomError(req.UserID, req.RoomID)
	}

	message, err := m.messageRepo.AddMessage(txContext, req.RoomID, req.UserID, req.Content)
	if err != nil {
		m.logger.Error(requestId, "m.messageRepo.AddMessage", req, err)
		tx.Rollback()
		return nil, m.errWarpper.NewDBServiceError(err)
	}

	err = tx.Commit().Error
	if err != nil {
		m.logger.Error(requestId, "tx.Commit", req, err)
		return nil, m.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.AddMessageResponse{
		ID:        message.ID,
		CreatedAt: common.TimeToUint64(message.CreatedAt),
	}, nil
}

func (m *messageServiceImpl) FetchMessages(ctx context.Context, req *dto.FetchMessageRequest) (*dto.FetchMessageResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	m.logger.Info(requestId, "start", req, nil)
	defer func() { m.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	InRoom, err := m.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		m.logger.Error(requestId, "m.roomRepo.CheckUserInRoom", req, err)
		return nil, m.errWarpper.NewDBServiceError(err)
	} else if !InRoom {
		tx.Rollback()
		return nil, m.errWarpper.NewUserNotInRoomError(req.UserID, req.RoomID)
	}

	messages, nextCursor, err := m.messageRepo.FetchMessages(txContext, req.RoomID, common.Uint64ToTime(req.TimeCursor), req.MessageSize)
	if err != nil {
		tx.Rollback()
		m.logger.Error(requestId, "m.messageRepo.FetchMessages", req, err)
		return nil, m.errWarpper.NewDBServiceError(err)
	}

	err = tx.Commit().Error
	if err != nil {
		m.logger.Error(requestId, "tx.Commit", req, err)
		return nil, m.errWarpper.NewDBCommitServiceError(err)
	}

	answer := &dto.FetchMessageResponse{
		NextTimeCursor: common.TimeToUint64(nextCursor),
	}
	messageResp := make([]dto.Message, len(messages))
	for i, message := range messages {
		messageResp[i].ID = message.ID
		messageResp[i].UserID = message.UserID
		messageResp[i].Content = message.Content
		messageResp[i].CreatedAt = common.TimeToUint64(message.CreatedAt)
	}
	answer.Messages = messageResp
	return answer, nil
}
