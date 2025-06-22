package service

import (
	"ChatRoomAPI/src/cache"
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
	messageRepo  repository.MessageRepository
	roomRepo     repository.RoomRepository
	stickerRepo  repository.StickerRepository
	errWarpper   dtoError.ServiceErrorWarpper
	logger       logger.Logger
	stickerCache cache.StickerCache
}

var message MessageService

func init() {
	message = &messageServiceImpl{
		messageRepo:  repository.GetMessageRepository(),
		roomRepo:     repository.GetRoomRepository(),
		errWarpper:   dtoError.GetServiceErrorWarpper(),
		logger:       logger.NewLogger(),
		stickerRepo:  repository.GetStickerRepository(),
		stickerCache: cache.GetStickerCache(),
	}
}

func GetMessageService() MessageService {
	return message
}

func (m *messageServiceImpl) checkStickerFromContent(ctx context.Context, content string, userId uint64) (string, error) {
	requestId := common.GetUUID(ctx)
	data := map[string]any{"userId": userId, "content": content}

	chunks := strings.Split(content, " ")
	userStickerSetCacheMap, err := m.stickerCache.GetAllStickerSetInfoByUser(ctx, userId)
	if err != nil {
		m.logger.Error(requestId, "m.stickerCache.GetAllStickerSetInfoByUser", data, err)
		m.stickerCache.ClearAllStickerCacheByUser(ctx, userId)
		stickerSetList, err := m.stickerRepo.GetAllAvailableStickersInfo(ctx, userId)
		if err != nil {
			return "", err
		}

		userStickerSetCacheMap = make(map[uint64]*cache.StickerSetCacheInfo)
		for _, stickerSet := range stickerSetList {
			stickerSetCacheInfo := &cache.StickerSetCacheInfo{
				Id:       stickerSet.Id,
				Name:     stickerSet.Name,
				Author:   stickerSet.Author,
				Price:    stickerSet.Price,
				Stickers: make(map[uint64]*cache.StickerCacheInfo),
			}

			for _, sticker := range stickerSet.Stickers {
				stickerSetCacheInfo.Stickers[sticker.Id] = &cache.StickerCacheInfo{
					Id:   sticker.Id,
					Name: sticker.Name,
				}
			}
			userStickerSetCacheMap[stickerSet.Id] = stickerSetCacheInfo
		}
		err = m.stickerCache.StoreStickerSetInfoByUser(ctx, userId, userStickerSetCacheMap)
		if err != nil {
			m.logger.Error(requestId, "m.stickerCache.StoreStickerSetInfoByUser", data, err)
		}
	}

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

		StickerSetCache, ok := userStickerSetCacheMap[stickerSetId]
		if !ok {
			chunks[i] = ""
		}

		_, ok = StickerSetCache.Stickers[stickerId]
		if !ok {
			chunks[i] = ""
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
