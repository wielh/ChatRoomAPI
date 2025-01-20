package service

import (
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/repository"
	"context"
)

type RoomAdminService interface {
	AdminChange(ctx context.Context, req *dto.AdminChangeRequest) (*dto.AdminChangeResponse, *dtoError.ServiceError)
	InviteNewUser(ctx context.Context, req *dto.InviteNewUserRequest) (*dto.InviteNewUserResponse, *dtoError.ServiceError)
	InviteNewUserCancel(ctx context.Context, req *dto.InviteNewUserCancelRequest) (*dto.InviteNewUserCancelResponse, *dtoError.ServiceError)
	FetchInvitationsByAdmin(ctx context.Context, req *dto.FetchInvitationByAdminRequest) (*dto.FetchInvitationByAdminResponse, *dtoError.ServiceError)
	ConfrimApply(ctx context.Context, req *dto.ConfrimApplyRequest) (*dto.ConfrimApplyResponse, *dtoError.ServiceError)
	FetchApplicationByAdmin(ctx context.Context, req *dto.FetchApplicationByAdminRequest) (*dto.FetchApplicationByAdminResponse, *dtoError.ServiceError)
	DeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (*dto.DeleteUserResponse, *dtoError.ServiceError)
}

type roomAdminServiceImpl struct {
	roomRepo        repository.RoomRepository
	applicationRepo repository.ApplicationRepository
	invitationRepo  repository.InvitationRepository
	errWarpper      dtoError.ServiceErrorWarpper
	userRepo        repository.AccountRepository
}

var roomAdmin RoomAdminService

func init() {
	roomAdmin = &roomAdminServiceImpl{
		roomRepo:        repository.GetRoomRepository(),
		applicationRepo: repository.GetApplicationRepository(),
		invitationRepo:  repository.GetInvitationRepository(),
		errWarpper:      dtoError.GetServiceErrorWarpper(),
		userRepo:        repository.GetAccountRepository(),
	}

}

func GetRoomAdminService() RoomAdminService {
	return roomAdmin
}

func (r *roomAdminServiceImpl) AdminChange(ctx context.Context, req *dto.AdminChangeRequest) (*dto.AdminChangeResponse, *dtoError.ServiceError) {
	if req.AdminUserID == req.UserID {
		return nil,
			&dtoError.ServiceError{
				StatusCode:     400,
				InternalError:  nil,
				ExtrenalReason: "AdminUserID and UserID are the same",
			}
	}

	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !roomExist {
		tx.Rollback()
		return nil, r.errWarpper.NewRoomNotExistError(req.RoomID)
	}

	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isAdmin {
		tx.Rollback()
		return nil, r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
	}

	isUser, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isUser {
		tx.Rollback()
		return nil, r.errWarpper.NewUserNotInRoomError(req.UserID, req.RoomID)
	}

	ok, err := r.roomRepo.AdminChange(txContext, req.RoomID, req.AdminUserID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return nil, r.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.AdminChangeResponse{}, nil
}

func (r *roomAdminServiceImpl) InviteNewUser(ctx context.Context, req *dto.InviteNewUserRequest) (*dto.InviteNewUserResponse, *dtoError.ServiceError) {
	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !roomExist {
		tx.Rollback()
		return nil, r.errWarpper.NewRoomNotExistError(req.RoomID)
	}

	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isAdmin {
		tx.Rollback()
		return nil, r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
	}

	userExist, err := r.userRepo.CheckUserExist(txContext, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !userExist {
		tx.Rollback()
		return nil, r.errWarpper.NewUserNotExist(req.UserID)
	}

	isUser, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if isUser {
		tx.Rollback()
		return nil, r.errWarpper.NewUserAlreadyInRoomError(req.UserID, req.RoomID)
	}

	repeat, err := r.invitationRepo.CheckInvitationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if repeat {
		tx.Rollback()
		return nil, r.errWarpper.NewUserIsInvitedError(req.UserID, req.RoomID)
	}

	err = r.invitationRepo.InviteNewUser(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.InviteNewUserResponse{}, nil
}

func (r *roomAdminServiceImpl) InviteNewUserCancel(ctx context.Context, req *dto.InviteNewUserCancelRequest) (*dto.InviteNewUserCancelResponse, *dtoError.ServiceError) {
	txContext, tx := repository.SetTxContext(ctx)
	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isAdmin {
		tx.Rollback()
		return nil, r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
	}

	invited, err := r.invitationRepo.CheckInvitationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !invited {
		tx.Rollback()
		return nil, r.errWarpper.NewUserIsNotInvitedError(req.UserID, req.RoomID)
	}

	ok, err := r.invitationRepo.InviteNewUserRequestDelete(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return nil, r.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &dto.InviteNewUserCancelResponse{}, nil
}

func (r *roomAdminServiceImpl) FetchInvitationsByAdmin(ctx context.Context, req *dto.FetchInvitationByAdminRequest) (*dto.FetchInvitationByAdminResponse, *dtoError.ServiceError) {
	txContext, tx := repository.SetTxContext(ctx)
	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isAdmin {
		tx.Rollback()
		return nil, r.errWarpper.NewNotAdminOfRoomError(req.AdminID, req.RoomID)
	}

	records, err := r.invitationRepo.FetchInvitationsByAdmin(txContext, req.AdminID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	}

	answer := dto.FetchInvitationByAdminResponse{RoomID: req.RoomID}
	answer.UserInfos = make([]dto.UserInfo, len(records))
	for i, record := range records {
		answer.UserInfos[i].UserID = record.UserId
		answer.UserInfos[i].Username = record.Name
		answer.UserInfos[i].Email = record.Email
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	return &answer, nil
}

func (r *roomAdminServiceImpl) ConfrimApply(ctx context.Context, req *dto.ConfrimApplyRequest) (*dto.ConfrimApplyResponse, *dtoError.ServiceError) {
	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !roomExist {
		tx.Rollback()
		r.invitationRepo.InviteNewUserRequestDelete(ctx, req.RoomID, req.UserID)
		return nil, r.errWarpper.NewRoomNotExistError(req.RoomID)
	}

	isAdmin, err := r.roomRepo.CheckAdminUserInRoom(txContext, req.RoomID, req.AdminUserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !isAdmin {
		tx.Rollback()
		return nil, r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
	}

	isUser, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if isUser {
		tx.Rollback()
		return nil, r.errWarpper.NewUserAlreadyInRoomError(req.UserID, req.RoomID)
	}

	applied, err := r.applicationRepo.CheckApplicationExist(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		return nil, r.errWarpper.NewDBServiceError(err)
	} else if !applied {
		tx.Rollback()
		return nil, r.errWarpper.NewUserNotApplyError(req.UserID, req.RoomID)
	}

	if req.Allowed {
		_, err = r.roomRepo.AddUser(txContext, req.RoomID, req.UserID)
		if err != nil {
			tx.Rollback()
			return nil, r.errWarpper.NewDBServiceError(err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, r.errWarpper.NewDBCommitServiceError(err)
	}
	go r.applicationRepo.RoomJoinApplyRequestDelete(txContext, req.RoomID, req.UserID)
	return &dto.ConfrimApplyResponse{}, nil
}

func (r *roomAdminServiceImpl) FetchApplicationByAdmin(ctx context.Context, req *dto.FetchApplicationByAdminRequest) (*dto.FetchApplicationByAdminResponse, *dtoError.ServiceError) {
	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
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
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !isAdmin {
		tx.Rollback()
		serviceErr := r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
		return nil, serviceErr
	}

	records, err := r.applicationRepo.FetchApplicationsByAdmin(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	}

	answer := dto.FetchApplicationByAdminResponse{RoomID: req.RoomID}
	answer.UserInfos = make([]dto.UserInfo, len(records))
	for i, record := range records {
		answer.UserInfos[i].UserID = record.UserId
		answer.UserInfos[i].Username = record.Name
		answer.UserInfos[i].Email = record.Email
	}

	err = tx.Commit().Error
	if err != nil {
		serviceErr := r.errWarpper.NewDBCommitServiceError(err)
		return nil, serviceErr
	}
	return &answer, nil
}

func (r *roomAdminServiceImpl) DeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (*dto.DeleteUserResponse, *dtoError.ServiceError) {

	txContext, tx := repository.SetTxContext(ctx)
	roomExist, err := r.roomRepo.RoomExist(txContext, req.RoomID)
	if err != nil {
		tx.Rollback()
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
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !isAdmin {
		tx.Rollback()
		serviceErr := r.errWarpper.NewNotAdminOfRoomError(req.AdminUserID, req.RoomID)
		return nil, serviceErr
	}

	isUser, err := r.roomRepo.CheckUserInRoom(txContext, req.RoomID, req.UserID)
	if err != nil {
		tx.Rollback()
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !isUser {
		tx.Rollback()
		serviceErr := r.errWarpper.NewUserNotInRoomError(req.UserID, req.RoomID)
		return nil, serviceErr
	}

	ok, err := r.roomRepo.DeleteUser(txContext, req.RoomID, req.AdminUserID, req.UserID)
	if err != nil {
		tx.Rollback()
		serviceErr := r.errWarpper.NewDBServiceError(err)
		return nil, serviceErr
	} else if !ok {
		tx.Rollback()
		serviceErr := r.errWarpper.NewDBNoAffectedServiceError()
		return nil, serviceErr
	}

	err = tx.Commit().Error
	if err != nil {
		serviceErr := r.errWarpper.NewDBCommitServiceError(err)
		return nil, serviceErr
	}
	return &dto.DeleteUserResponse{}, nil
}
