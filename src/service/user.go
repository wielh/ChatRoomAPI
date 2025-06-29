package service

import (
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/dto"
	"ChatRoomAPI/src/dtoError"
	"ChatRoomAPI/src/logger"
	"ChatRoomAPI/src/repository"
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	UserRegisterService(ctx context.Context, req *dto.UserRegisterRequest) (*dto.UserRegisterResponse, *dtoError.ServiceError)
	UserLoginService(ctx context.Context, req *dto.UserLoginRequest) (*dto.UserLoginResponse, *dtoError.ServiceError)
	ResetPasswordService(ctx context.Context, req *dto.ResetPasswordRequest) *dtoError.ServiceError
	UserInfoService(ctx context.Context, req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, *dtoError.ServiceError)
}

type userServiceImpl struct {
	logger      logger.Logger
	accountRepo repository.AccountRepository
	errWarpper  dtoError.ServiceErrorWarpper
	tracer      trace.Tracer
}

var user UserService

func init() {
	user = &userServiceImpl{
		accountRepo: repository.GetAccountRepository(),
		errWarpper:  dtoError.GetServiceErrorWarpper(),
		logger:      logger.NewInfoLogger(),
		tracer:      otel.Tracer("userService"),
	}
}

func GetAccountService() UserService {
	return user
}

func (a *userServiceImpl) UserRegisterService(ctx context.Context, req *dto.UserRegisterRequest) (*dto.UserRegisterResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	ctx, span := a.tracer.Start(ctx, "Register")
	defer span.End()

	data := map[string]any{
		"username": req.Username,
		"name":     req.Name,
		"email":    req.Email,
	}
	a.logger.Info(requestId, "start", data, nil)

	hashedPassword, _ := hashPassword(req.Password)
	parsedTime, _ := time.Parse("2006-01-02", req.Birthday)
	userModel, ok, err := a.accountRepo.UserRegister(ctx, req.Username, hashedPassword, req.Name, req.Email, parsedTime)
	if err != nil {
		a.logger.Error(requestId, "a.accountRepo.UserRegister", data, err)
		return nil, a.errWarpper.NewDBServiceError(err)
	} else if !ok {
		return nil, a.errWarpper.NewUserHasRegisterdError(req.Username)
	}
	return &dto.UserRegisterResponse{ID: userModel.Id}, nil
}

func (a *userServiceImpl) UserLoginService(ctx context.Context, req *dto.UserLoginRequest) (*dto.UserLoginResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	ctx, span := a.tracer.Start(ctx, "Login")
	defer span.End()

	data := map[string]any{"username": req.Username}
	a.logger.Info(requestId, "start", data, nil)

	userModel, exist, err := a.accountRepo.SelectUserByName(ctx, req.Username)
	if err != nil {
		a.logger.Error(requestId, "a.accountRepo.SelectUserByName", data, err)
		return nil, a.errWarpper.NewDBServiceError(err)
	} else if !exist {
		return nil, a.errWarpper.NewLoginFailedServiceError(nil)
	}

	err = comparePassword(userModel.Password, req.Password)
	if err != nil {
		return nil, a.errWarpper.NewLoginFailedServiceError(err)
	}

	return &dto.UserLoginResponse{
		ID:       userModel.Id,
		Username: userModel.Username,
	}, nil
}

func (a *userServiceImpl) ResetPasswordService(ctx context.Context, req *dto.ResetPasswordRequest) *dtoError.ServiceError {
	requestId := common.GetUUID(ctx)
	ctx, span := a.tracer.Start(ctx, "ResetPassword")
	defer span.End()

	data := map[string]any{"username": req.Username}
	a.logger.Info(requestId, "start", data, nil)

	txContext, tx := repository.SetTxContext(ctx)
	user, ok, err := a.accountRepo.SelectUserByName(txContext, req.Username)
	if err != nil {
		a.logger.Error(requestId, "a.accountRepo.SelectUserByName", data, err)
		tx.Rollback()
		return a.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return a.errWarpper.NewRessetPasswordServiceError(err)
	}

	err = comparePassword(user.Password, req.Password)
	if err != nil {
		tx.Rollback()
		return a.errWarpper.NewRessetPasswordServiceError(err)
	}

	newHashPassword, _ := hashPassword(req.NewPassword)
	ok, err = a.accountRepo.UpdatePassword(txContext, user.Id, newHashPassword)
	if err != nil {
		a.logger.Error(requestId, "a.accountRepo.UpdatePassword", data, err)
		tx.Rollback()
		return a.errWarpper.NewDBServiceError(err)
	} else if !ok {
		tx.Rollback()
		return a.errWarpper.NewDBNoAffectedServiceError()
	}

	err = tx.Commit().Error
	if err != nil {
		a.logger.Error(requestId, "tx.Commit", data, err)
		return a.errWarpper.NewDBCommitServiceError(err)
	}
	return nil
}

func (a *userServiceImpl) UserInfoService(ctx context.Context, req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, *dtoError.ServiceError) {
	requestId := common.GetUUID(ctx)
	data := map[string]any{"id": req.ID}
	a.logger.Info(requestId, "start", data, nil)

	user, err := a.accountRepo.UserInfo(ctx, req.ID)
	if err != nil {
		a.logger.Error(requestId, "a.accountRepo.UserInfo", data, err)
		return nil, a.errWarpper.NewDBServiceError(err)
	}

	return &dto.GetUserInfoResponse{
		Id:       user.Id,
		Username: user.Username,
		Name:     user.Name,
		Birthday: user.Birthday.Format("2006-01-02"),
		Email:    user.Email,
	}, nil
}

// ====================================================================================

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func comparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
