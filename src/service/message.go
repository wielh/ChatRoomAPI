package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
	"strconv"
	"strings"
)

type MessageService interface {
	AddMessage(ctx context.Context, req *dto.AddMessageRequest) (*dto.AddMessageResponse, *dtoError.ServiceError)
	FetchMessages(ctx context.Context, req *dto.FetchMessageRequest) (*dto.FetchMessageResponse, *dtoError.ServiceError)
}

type messageServiceImpl struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
	stickerRepo repository.StickerRepository
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
		stickerRepo: repository.GetStickerRepository(),
	}
}

func GetMessageService() MessageService {
	return message
}

func (m *messageServiceImpl) checkStickerFromContent(ctx context.Context, content string, userId uint64) (string, error) {
	chunks := strings.Split(content, " ")
	for i, chunk := range chunks {
		subchunks := strings.Split(chunk, "::")
		if len(subchunks) != 3 {
			continue
		}
		if subchunks[0] != "sticker" {
			continue
		}

		stickerSetId, err := strconv.ParseUint(subchunks[1], 10, 64)
		if err != nil {
			continue
		}

		stickerId, err := strconv.ParseUint(subchunks[2], 10, 64)
		if err != nil {
			continue
		}

		_, exist, err := m.stickerRepo.CheckAvailable(ctx, userId, stickerSetId, stickerId)
		if !exist {
			chunks[i] = ""
		} else if err != nil {
			return "", err
		}
	}

	newContent := strings.Join(chunks, " ")
	return newContent, nil
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

	newContent, err := m.checkStickerFromContent(ctx, req.Content, req.UserID)
	if err != nil {
		tx.Rollback()
		m.logger.Error(requestId, "m.checkStickerFromContent", req, err)
		return nil, m.errWarpper.NewDBServiceError(err)
	}

	message, err := m.messageRepo.AddMessage(txContext, req.RoomID, req.UserID, newContent)
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
