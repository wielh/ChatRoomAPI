package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
)

type RoomUserService interface {
	ConfrimInvite(ctx context.Context, req *dto.ConfrimInviteRequest) (*dto.ConfrimInviteResponse, *dtoError.ServiceError)
	FetchInvitationsByUser(ctx context.Context, req *dto.FetchInvitationByUserRequest) (*dto.FetchInvitationByUserResponse, *dtoError.ServiceError)
	RoomJoinApply(ctx context.Context, req *dto.RoomJoinApplyRequest) (*dto.RoomJoinApplyResponse, *dtoError.ServiceError)
	RoomJoinApplyCancel(ctx context.Context, req *dto.RoomJoinApplyCancelRequest) (*dto.RoomJoinApplyCancelResponse, *dtoError.ServiceError)
	FetchApplicationByUser(ctx context.Context, req *dto.FetchApplicationByUserRequest) (*dto.FetchApplicationByUserResponse, *dtoError.ServiceError)
}

type roomUserServiceImpl struct {
	roomRepo        repository.RoomRepository
	applicationRepo repository.ApplicationRepository
	invitationRepo  repository.InvitationRepository
	errWarpper      dtoError.ServiceErrorWarpper
	logger          logger.Logger
}

var roomUser RoomUserService

func init() {
	roomUser = &roomUserServiceImpl{
		roomRepo:        repository.GetRoomRepository(),
		applicationRepo: repository.GetApplicationRepository(),
		invitationRepo:  repository.GetInvitationRepository(),
		errWarpper:      dtoError.GetServiceErrorWarpper(),
		logger:          logger.NewLogger(),
	}
}

func GetRoomUserService() RoomUserService {
	return roomUser
}

func (r *roomUserServiceImpl) ConfrimInvite(ctx context.Context, req *dto.ConfrimInviteRequest) (*dto.ConfrimInviteResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.roomRepo.RoomExist", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !roomExist {
		tx.Rollback()
		return nil, r.errWarpper.NewRoomNotExistError(req.RoomID)
	}

	InRoom, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.roomRepo.CheckUserInRoom", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if InRoom {
		tx.Rollback()
		return nil, r.errWarpper.NewUserAlreadyInRoomError(req.UserID, req.RoomID)
	}

	exist, err := r.invitationRepo.CheckInvitationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.invitationRepo.CheckInvitationExist", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !exist {
		tx.Rollback()
		return nil, r.errWarpper.NewUserIsNotInvitedError(req.UserID, req.RoomID)
	}

	if req.Allowed {
		_, err = r.roomRepo.AddUser(txContext, req.RoomID, req.UserID)
		if err != nil {
			r.logger.Error(requestId, "r.roomRepo.AddUser", req, err)
			tx.Rollback()
			return nil, r.errWarpper.NewDBServiceError(err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		r.logger.Error(requestId, "tx.Commit", req, err)
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}

	go r.invitationRepo.InviteNewUserRequestDelete(ctx, req.RoomID, req.UserID)
	return &dto.ConfrimInviteResponse{}, nil
}

func (r *roomUserServiceImpl) FetchInvitationsByUser(ctx context.Context, req *dto.FetchInvitationByUserRequest) (*dto.FetchInvitationByUserResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	skip, pageSize := GetSkip(int(req.Page), int(req.PageSize))
	records, err := r.invitationRepo.FetchInvitationsByUser(ctx, req.UserID, skip, pageSize)
	if err != nil {
		r.logger.Error(requestId, "r.invitationRepo.FetchInvitationsByUser", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	answer := dto.FetchInvitationByUserResponse{UserID: req.UserID}
	answer.RoomInfos = make([]dto.RoomInfo, len(records))
	for i, record := range records {
		answer.RoomInfos[i].AdminID = record.AdminUserId
		answer.RoomInfos[i].Description = record.Description
		answer.RoomInfos[i].RoomID = record.RoomId
		answer.RoomInfos[i].RoomName = record.RoomName
		answer.RoomInfos[i].UserIDs = common.PQInt64ArrayToUInt64Array(record.UserIds)
	}
	return &answer, nil
}

func (r *roomUserServiceImpl) RoomJoinApply(ctx context.Context, req *dto.RoomJoinApplyRequest) (*dto.RoomJoinApplyResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		r.logger.Error(requestId, "r.roomRepo.RoomExist", req, err)
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !roomExist {
		tx.Rollback()
		return nil, r.errWarpper.NewRoomNotExistError(req.RoomID)
	}

	isUser, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		r.logger.Error(requestId, "r.roomRepo.CheckUserInRoom", req, err)
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if isUser {
		tx.Rollback()
		return nil, r.errWarpper.NewUserAlreadyInRoomError(req.UserID, req.RoomID)
	}

	exist, err := r.applicationRepo.CheckApplicationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		r.logger.Error(requestId, "r.applicationRepo.CheckApplicationExist", req, err)
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if exist {
		tx.Rollback()
		return nil, r.errWarpper.NewUserApplyError(req.UserID, req.RoomID)
	}

	ok, err := r.applicationRepo.RoomJoinApply(txContext, req.RoomID, req.UserID)
	if err != nil {
		r.logger.Error(requestId, "r.applicationRepo.RoomJoinApply", req, err)
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return nil, r.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		r.logger.Error(requestId, "tx.Commit", req, err)
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.RoomJoinApplyResponse{}, nil
}

func (r *roomUserServiceImpl) RoomJoinApplyCancel(ctx context.Context, req *dto.RoomJoinApplyCancelRequest) (*dto.RoomJoinApplyCancelResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	exist, err := r.applicationRepo.CheckApplicationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.applicationRepo.CheckApplicationExist", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !exist {
		tx.Rollback()
		return nil, r.errWarpper.NewUserNotApplyError(req.UserID, req.RoomID)
	}

	ok, err := r.applicationRepo.RoomJoinApplyRequestDelete(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.applicationRepo.RoomJoinApplyRequestDelete", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return nil, r.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		r.logger.Error(requestId, "tx.Commit", req, err)
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.RoomJoinApplyCancelResponse{}, nil
}

func (r *roomUserServiceImpl) FetchApplicationByUser(ctx context.Context, req *dto.FetchApplicationByUserRequest) (*dto.FetchApplicationByUserResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	skip, pageSize := GetSkip(int(req.Page), int(req.PageSize))
	records, err := r.applicationRepo.FetchApplicationsByUser(ctx, req.UserID, skip, pageSize)
	if err != nil {
		r.logger.Error(requestId, "r.applicationRepo.FetchApplicationsByUser", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	answer := dto.FetchApplicationByUserResponse{UserID: req.UserID}
	answer.RoomInfos = make([]dto.RoomInfo, len(records))
	for i, record := range records {
		answer.RoomInfos[i].AdminID = record.AdminUserId
		answer.RoomInfos[i].Description = record.Description
		answer.RoomInfos[i].RoomID = record.Id
		answer.RoomInfos[i].RoomName = record.RoomName
		answer.RoomInfos[i].UserIDs = common.PQInt64ArrayToUInt64Array(record.UserIds)
	}
	return &answer, nil
}
