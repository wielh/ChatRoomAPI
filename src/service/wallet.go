package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type WalletService interface {
	GetState(ctx context.Context, req *dto.GetStateRequest) (*dto.GetStateResponse, *dtoError.ServiceError)
	Charge(ctx context.Context, req *dto.ChargeRequest) (*dto.ChargeResponse, *dtoError.ServiceError)
}

type walletServiceImpl struct {
	logger             logger.Logger
	walletRepository   repository.WalletRepository
	errWarpper         dtoError.ServiceErrorWarpper
	tracer             trace.Tracer
	MIN_CHARGE_ACCOUNT uint32
	MAX_CHARGE_ACCOUNT uint32
}

var wallet WalletService

func init() {
	wallet = &walletServiceImpl{
		walletRepository:   repository.GetWalletRepository(),
		errWarpper:         dtoError.GetServiceErrorWarpper(),
		logger:             logger.NewErrorLogger(),
		tracer:             otel.Tracer("walletService"),
		MIN_CHARGE_ACCOUNT: 1,
		MAX_CHARGE_ACCOUNT: 1000,
	}
}

func GetWalletService() WalletService {
	return wallet
}

func (w *walletServiceImpl) GetState(ctx context.Context, req *dto.GetStateRequest) (*dto.GetStateResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	wallet, exist, err := w.walletRepository.GetState(ctx, req.UserID)
	if err != nil {
		w.logger.Error(requestId, "w.walletRepository.GetState", req, err)
		return nil, w.errWarpper.NewDBServiceError(err)
	} else if !exist {
		return nil, w.errWarpper.NewUserNotChargedError(req.UserID)
	}
	return &dto.GetStateResponse{Money: wallet.Money}, nil
}

// TODO: this is mock charge api endpoint only
func (w *walletServiceImpl) Charge(ctx context.Context, req *dto.ChargeRequest) (*dto.ChargeResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	ctx, span := w.tracer.Start(ctx, "Charge")
	defer span.End()

	w.logger.Info(requestId, "charge request start", req, nil)
	if req.Money < w.MIN_CHARGE_ACCOUNT || req.Money > w.MAX_CHARGE_ACCOUNT {
		return nil, w.errWarpper.NewUserChargeMoneyExcessError(req.UserID, req.Money, w.MIN_CHARGE_ACCOUNT, w.MAX_CHARGE_ACCOUNT)
	}

	txContext, tx := repository.SetTxContext(ctx)
	err := w.walletRepository.WalletInit(txContext, req.UserID)
	if err != nil {
		w.logger.Error(requestId, "w.walletRepository.WalletInit", req, err)
		tx.Rollback()
		return nil, w.errWarpper.NewDBServiceError(err)
	}

	err = w.walletRepository.Charge(txContext, req.UserID, req.Money)
	if err != nil {
		w.logger.Error(requestId, "w.walletRepository.Charge", req, err)
		tx.Rollback()
		return nil, w.errWarpper.NewDBServiceError(err)
	}

	err = tx.Commit().Error
	if err != nil {
		w.logger.Error(requestId, "tx.Commit", req, err)
		return nil, w.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.ChargeResponse{
		OK:         true,
		MinAccount: w.MIN_CHARGE_ACCOUNT,
		MaxAccount: w.MAX_CHARGE_ACCOUNT,
	}, nil
}
