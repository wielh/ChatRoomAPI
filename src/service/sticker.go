package service

import (
	"ChatRoomAPI/src/cache"
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type StickerService interface {
	GetStickerSetInfo(ctx context.Context, req *dto.GetStickerSetInfoRequest) (*dto.GetStickerSetInfoResponse, *dtoError.ServiceError)
	BuyStickerSet(ctx context.Context, req *dto.BuyStickerSetRequest) (*dto.BuyStickerResponse, *dtoError.ServiceError)
	GetAllAvailableStickersInfo(ctx context.Context, req *dto.GetAllAvailableStickersInfoRequest) (*dto.GetAllAvailableStickersInfoResponse, *dtoError.ServiceError)
}

type stickerServiceImpl struct {
	logger           logger.Logger
	walletRepository repository.WalletRepository
	stickerRepo      repository.StickerRepository
	errWarpper       dtoError.ServiceErrorWarpper
	stickerCache     cache.StickerCache
	tracer           trace.Tracer
}

var sticker StickerService

func init() {
	sticker = &stickerServiceImpl{
		stickerRepo:      repository.GetStickerRepository(),
		walletRepository: repository.GetWalletRepository(),
		errWarpper:       dtoError.GetServiceErrorWarpper(),
		logger:           logger.NewInfoLogger(),
		stickerCache:     cache.GetStickerCache(),
		tracer:           otel.Tracer("stickerService"),
	}
}

func GetStickerService() StickerService {
	return sticker
}

func (s *stickerServiceImpl) GetStickerSetInfo(ctx context.Context, req *dto.GetStickerSetInfoRequest) (*dto.GetStickerSetInfoResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	stickerInfo, exist, err := s.stickerRepo.GetStickerSetInfo(ctx, req.StickerSetID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetStickerSetInfo", req, err)
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if !exist {
		return nil, s.errWarpper.NewStickerSetNotExistError(req.StickerSetID)
	}

	resp := dto.GetStickerSetInfoResponse{}
	resp.Id = stickerInfo.Id
	resp.Author = stickerInfo.Author
	resp.Name = stickerInfo.Name
	resp.Price = stickerInfo.Price
	resp.Stickers = make([]dto.StickerInfo, len(stickerInfo.Stickers))
	for i, item := range stickerInfo.Stickers {
		resp.Stickers[i].Id = item.Id
		resp.Stickers[i].Name = item.Name
	}
	return &resp, nil

}

func (s *stickerServiceImpl) BuyStickerSet(ctx context.Context, req *dto.BuyStickerSetRequest) (*dto.BuyStickerResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	ctx, span := s.tracer.Start(ctx, "BuyStickerSet")
	defer span.End()
	s.logger.Info(requestId, "start BuyStickerSet", req, nil)

	txContext, tx := repository.SetTxContext(ctx)
	stickerSet, exist, err := s.stickerRepo.GetStickerSetInfo(txContext, req.StickerSetId)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetStickerSetInfo", req, err)
		tx.Rollback()
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if !exist {
		tx.Rollback()
		return nil, s.errWarpper.NewStickerSetNotExistError(req.StickerSetId)
	}

	_, ok, err := s.walletRepository.Cost(txContext, req.UserID, stickerSet.Price)
	if err != nil {
		s.logger.Error(requestId, "s.walletRepository.Cost", req, err)
		tx.Rollback()
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return nil, s.errWarpper.NewUserMoneyNotEnoughError(req.UserID)
	}

	exist, err = s.stickerRepo.CheckStickerUserMappingExist(ctx, req.StickerSetId, req.UserID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetStickerSetInfo", req, err)
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if exist {
		return nil, s.errWarpper.NewStickerAlreadyBuyError(req.StickerSetId, req.UserID)
	}

	err = s.stickerRepo.StickerSetBindingToUser(txContext, req.StickerSetId, req.UserID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.StickerSetBindingToUser", req, err)
		tx.Rollback()
		return nil, s.errWarpper.NewDBServiceError(err)
	}

	detail := fmt.Sprintf("{Item: sticker,Id: %d, Name: %s, Price: %d}", stickerSet.Id, stickerSet.Name, stickerSet.Price)
	_, err = s.walletRepository.WriteLog(ctx, req.UserID, 1, stickerSet.Price, detail)
	if err != nil {
		s.logger.Error(requestId, "s.walletRepository.WriteLog", req, err)
		tx.Rollback()
		return nil, s.errWarpper.NewDBServiceError(err)
	}

	err = tx.Commit().Error
	if err != nil {
		s.logger.Error(requestId, "tx.Commit", req, err)
		return nil, s.errWarpper.NewDBCommitServiceError(err)
	}

	stickerCacheInfo := cache.StickerSetCacheInfo{
		Id:       stickerSet.Id,
		Name:     stickerSet.Name,
		Author:   stickerSet.Name,
		Price:    stickerSet.Price,
		Stickers: make(map[uint64]*cache.StickerCacheInfo),
	}
	for _, sticker := range stickerSet.Stickers {
		stickerCacheInfo.Stickers[sticker.Id] = &cache.StickerCacheInfo{
			Id: sticker.Id, Name: sticker.Name,
		}
	}

	s.stickerCache.InsertNewStickerSetInfoByUser(ctx, req.UserID, &stickerCacheInfo)
	return &dto.BuyStickerResponse{}, nil
}

func (s *stickerServiceImpl) GetAllAvailableStickersInfo(ctx context.Context, req *dto.GetAllAvailableStickersInfoRequest) (*dto.GetAllAvailableStickersInfoResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)

	stickerSetCacheMap, keyExist, err := s.stickerCache.GetAllStickerSetInfoByUser(ctx, req.UserID)
	if err == nil && keyExist {
		stickerSetInfoList := []*dto.StickerSetInfo{}
		for _, stickerSet := range stickerSetCacheMap {
			info := dto.StickerSetInfo{
				Id:     stickerSet.Id,
				Name:   stickerSet.Name,
				Author: stickerSet.Author,
				Price:  stickerSet.Price,
			}
			stickerInfoList := []dto.StickerInfo{}
			for _, sticker := range stickerSet.Stickers {
				stickerInfoList = append(stickerInfoList,
					dto.StickerInfo{
						Id:   sticker.Id,
						Name: sticker.Name,
					},
				)
			}
			info.Stickers = stickerInfoList
			stickerSetInfoList = append(stickerSetInfoList, &info)
		}
		return &dto.GetAllAvailableStickersInfoResponse{StickerSetInfoList: stickerSetInfoList}, nil
	}

	s.logger.Error(requestId, "s.stickerCache.GetAllStickerSetInfoByUser", req, err)
	err = s.stickerCache.ClearAllStickerCacheByUser(ctx, req.UserID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerCache.ClearAllStickerCacheByUser", req, err)
	}

	stickerSetList, err := s.stickerRepo.GetAllAvailableStickersInfo(ctx, req.UserID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetAllAvailableStickersInfo", req, err)
		return nil, s.errWarpper.NewDBServiceError(err)
	}

	go func() {
		cacheInfos := make(map[uint64]*cache.StickerSetCacheInfo)
		for _, stickerSet := range stickerSetList {
			cacheInfos[stickerSet.Id] = &cache.StickerSetCacheInfo{
				Id:     stickerSet.Id,
				Name:   stickerSet.Name,
				Author: stickerSet.Author,
				Price:  stickerSet.Price,
			}

			stickerInfoMap := make(map[uint64]*cache.StickerCacheInfo)
			for _, sticker := range stickerSet.Stickers {
				stickerInfoMap[sticker.Id] = &cache.StickerCacheInfo{
					Id:   sticker.Id,
					Name: sticker.Name,
				}
			}
			cacheInfos[stickerSet.Id].Stickers = stickerInfoMap
		}
		err = s.stickerCache.StoreStickerSetInfoByUser(ctx, req.UserID, cacheInfos)
		if err != nil {
			s.logger.Error(requestId, "s.stickerCache.StoreStickerSetInfoByUser", req, err)
		}
	}()

	response := dto.GetAllAvailableStickersInfoResponse{}
	stickerSetInfoList := make([]*dto.StickerSetInfo, len(stickerSetList))
	for i, stickerSet := range stickerSetList {
		stickerSetInfoList[i] = &dto.StickerSetInfo{
			Id:     stickerSet.Id,
			Name:   stickerSet.Name,
			Author: stickerSet.Author,
			Price:  stickerSet.Price,
		}

		stickerInfoList := make([]dto.StickerInfo, len(stickerSet.Stickers))
		for j, sticker := range stickerSet.Stickers {
			stickerInfoList[j] = dto.StickerInfo{
				Id:   sticker.Id,
				Name: sticker.Name,
			}
		}
		stickerSetInfoList[i].Stickers = stickerInfoList
	}
	response.StickerSetInfoList = stickerSetInfoList
	return &response, nil
}
