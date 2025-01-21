package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req *dto.CreateRoomRequest) (*dto.CreateRoomResponse, *dtoError.ServiceError)
	GetAvailbleRooms(ctx context.Context, req *dto.GetAvailbleRoomsRequest) (*dto.GetAvailbleRoomsResponse, *dtoError.ServiceError)
	ReadRoomInfo(ctx context.Context, req *dto.ReadRoomInfoRequest) (*dto.ReadRoomInfoResponse, *dtoError.ServiceError)
	DeleteRoom(ctx context.Context, req *dto.DeleteRoomRequest) (*dto.DeleteRoomResponse, *dtoError.ServiceError)
}

type roomServiceImpl struct {
	roomRepo   repository.RoomRepository
	errWarpper dtoError.ServiceErrorWarpper
	logger     logger.Logger
}

var room RoomService

func init() {
	room = &roomServiceImpl{
		roomRepo:   repository.GetRoomRepository(),
		errWarpper: dtoError.GetServiceErrorWarpper(),
		logger:     logger.NewLogger(),
	}
}

func GetRoomService() RoomService {
	return room
}

func (r *roomServiceImpl) CreateRoom(ctx context.Context, req *dto.CreateRoomRequest) (*dto.CreateRoomResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	roomInfo, ok, err := r.roomRepo.CreateRoom(ctx, req.UserID, req.RoomName, req.Description)
	if err != nil {
		r.logger.Error(requestId, "r.roomRepo.CreateRoom", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		return nil, r.errWarpper.NewUserHasRegisterdError(req.RoomName)
	}
	return &dto.CreateRoomResponse{RoomID: roomInfo.Id}, nil
}

func (r *roomServiceImpl) GetAvailbleRooms(ctx context.Context, req *dto.GetAvailbleRoomsRequest) (*dto.GetAvailbleRoomsResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	roomsInfo, err := r.roomRepo.GetAvailbleRooms(ctx, req.UserID)
	if err != nil {
		r.logger.Error(requestId, "r.roomRepo.GetAvailbleRooms", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	answer := make([]dto.ReadRoomInfoResponse, len(roomsInfo))
	for i, info := range roomsInfo {
		answer[i].ID = info.Id
		answer[i].Name = info.Name
		answer[i].AdminUserID = info.AdminUserID
		var uid []uint64
		for _, userid := range info.UserIDs {
			uid = append(uid, uint64(userid))
		}
		answer[i].UserIDs = uid
		answer[i].Description = info.Description
	}
	return &dto.GetAvailbleRoomsResponse{RoomsInfos: answer}, nil
}

func (r *roomServiceImpl) ReadRoomInfo(ctx context.Context, req *dto.ReadRoomInfoRequest) (*dto.ReadRoomInfoResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	room, err := r.roomRepo.ReadRoomInfo(ctx, req.RoomID)
	if err != nil {
		r.logger.Info(requestId, "r.roomRepo.ReadRoomInfo", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	var uid []uint64
	for _, userid := range room.UserIDs {
		uid = append(uid, uint64(userid))
	}

	return &dto.ReadRoomInfoResponse{
		ID:          req.UserID,
		Name:        room.Name,
		AdminUserID: room.AdminUserID,
		UserIDs:     uid,
		Description: room.Description,
	}, nil
}

func (r *roomServiceImpl) DeleteRoom(ctx context.Context, req *dto.DeleteRoomRequest) (*dto.DeleteRoomResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	r.logger.Info(requestId, "start", req, nil)
	defer func() { r.logger.Info(requestId, "end", req, nil) }()

	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.roomRepo.RoomExist", req, err)
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !roomExist {
		tx.Rollback()
		serviceErr := r.errWarpper.NewRoomNotExistError(req.RoomID)
		return nil, serviceErr
	}

	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		tx.Rollback()
		r.logger.Error(requestId, "r.roomRepo.CheckAdminUserInRoom", req, err)
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !isAdmin {
		tx.Rollback()
		serviceErr := r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
		return nil, serviceErr
	}

	ok, err := r.roomRepo.DeleteRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		r.logger.Error(requestId, "r.roomRepo.DeleteRoom", req, err)
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		return nil, r.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		r.logger.Error(requestId, "tx.Commit", req, err)
		serviceErr := r.errWarpper.NewDBCommitServiceError(err)
		return nil, serviceErr
	}
	return &dto.DeleteRoomResponse{}, nil
}
