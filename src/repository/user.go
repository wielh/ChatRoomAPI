package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type AccountRepository interface {
	UserRegister(ctx context.Context, username string, password string, name string, email string, birthday time.Time) (*model.User, bool, error)
	SelectUserByName(ctx context.Context, username string) (*model.User, bool, error)
	UpdatePassword(ctx context.Context, ID uint64, newHashedPassword string) (ok bool, err error)
	UserInfo(ctx context.Context, ID uint64) (*model.User, error)
	CheckUserExist(ctx context.Context, ID uint64) (exist bool, err error)
}

type accountRepositoryImpl struct {
	DB     *gorm.DB
	tracer trace.Tracer
}

var account AccountRepository

func init() {
	account = &accountRepositoryImpl{
		DB:     src.GlobalConfig.DB,
		tracer: otel.Tracer("accountRepository"),
	}
}

func GetAccountRepository() AccountRepository {
	return account
}

func (a *accountRepositoryImpl) UserRegister(ctx context.Context,
	username string, password string, name string, email string, birthday time.Time) (*model.User, bool, error) {
	tx := GetTxContext(ctx, a.DB)
	ctx, span := a.tracer.Start(ctx, "UserRegister")
	defer span.End()

	user := model.User{
		Username: username,
		Password: password,
		Name:     name,
		Email:    email,
		Birthday: birthday,
	}

	result := tx.Where("username=?", username).FirstOrCreate(&user)
	if result.Error != nil {
		return nil, false, result.Error
	}
	return &user, result.RowsAffected > 0, nil
}

func (a *accountRepositoryImpl) UserInfo(ctx context.Context, ID uint64) (*model.User, error) {
	tx := GetTxContext(ctx, a.DB)
	var user = model.User{Id: ID}
	result := tx.Select("Id", "Username", "Name", "Birthday", "Email").First(&user, "ID=?", ID)
	return &user, result.Error
}

func (a *accountRepositoryImpl) SelectUserByName(ctx context.Context, username string) (*model.User, bool, error) {
	tx := GetTxContext(ctx, a.DB)
	var user = model.User{Username: username}
	result := tx.Select("id", "username", "password").Where("username=?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, result.Error
	}
	return &user, true, nil
}

func (a *accountRepositoryImpl) UpdatePassword(ctx context.Context, ID uint64, newHashedPassword string) (bool, error) {
	tx := GetTxContext(ctx, a.DB)
	ctx, span := a.tracer.Start(ctx, "UpdatePassword")
	defer span.End()

	result := tx.Model(&model.User{}).Where("id=?", ID).Updates(map[string]interface{}{"password": newHashedPassword})
	if result.Error != nil {
		return false, result.Error
	} else if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (a *accountRepositoryImpl) CheckUserExist(ctx context.Context, ID uint64) (exist bool, err error) {
	tx := GetTxContext(ctx, a.DB)
	var user model.User
	result := tx.Select("username").Where("id=?", ID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
