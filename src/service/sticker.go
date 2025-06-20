package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
	"fmt"
)

type StickerService interface {
	GetStickerSetInfo(ctx context.Context, req *dto.GetStickerSetInfoRequest) (*dto.GetStickerSetInfoResponse, *dtoError.ServiceError)
	BuyStickerSet(ctx context.Context, req *dto.BuyStickerSetRequest) (*dto.BuyStickerResponse, *dtoError.ServiceError)
	CheckStickerSetAvailable(ctx context.Context, req *dto.CheckStickerSetAvailableRequest) (*dto.CheckStickerSetAvailableResponse, *dtoError.ServiceError)
}

type stickerServiceImpl struct {
	logger           logger.Logger
	walletRepository repository.WalletRepository
	stickerRepo      repository.StickerRepository
	errWarpper       dtoError.ServiceErrorWarpper
}

var sticker StickerService

func init() {
	sticker = &stickerServiceImpl{
		stickerRepo:      repository.GetStickerRepository(),
		walletRepository: repository.GetWalletRepository(),
		errWarpper:       dtoError.GetServiceErrorWarpper(),
		logger:           logger.NewZeroLogger(),
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
	return &dto.BuyStickerResponse{}, nil
}

func (s *stickerServiceImpl) CheckStickerSetAvailable(ctx context.Context, req *dto.CheckStickerSetAvailableRequest) (*dto.CheckStickerSetAvailableResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	_, exist, err := s.stickerRepo.GetStickerSetInfo(ctx, req.StickerSetID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetStickerSetInfo", req, err)
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if !exist {
		return &dto.CheckStickerSetAvailableResponse{Ok: false}, nil
	}

	exist, err = s.stickerRepo.CheckStickerUserMappingExist(ctx, req.StickerSetID, req.UserID)
	if err != nil {
		s.logger.Error(requestId, "s.stickerRepo.GetStickerSetInfo", req, err)
		return nil, s.errWarpper.NewDBServiceError(err)
	} else if !exist {
		return &dto.CheckStickerSetAvailableResponse{Ok: false}, nil
	}
	return &dto.CheckStickerSetAvailableResponse{Ok: true}, nil
}
