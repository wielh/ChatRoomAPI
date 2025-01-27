package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ApplicationJoinRoomRepository interface {
	FetchApplicationsByAdmin(ctx context.Context, roomID uint64, skip int, pageSize int) (record []*model.UserJoinApplyRecord, err error)
}

type ApplicationJoinUserRepository interface {
	FetchApplicationsByUser(ctx context.Context, userID uint64, skip int, pageSize int) (record []*model.RoomJoinApplyRecord, err error)
}

type ApplicationBaseRepository interface {
	RoomJoinApply(ctx context.Context, roomID uint64, userID uint64) (bool, error)
	RoomJoinApplyRequestDelete(ctx context.Context, roomID uint64, userID uint64) (bool, error)
	CheckApplicationExist(ctx context.Context, roomID uint64, userID uint64) (exist bool, err error)
}

type ApplicationRepository interface {
	ApplicationJoinRoomRepository
	ApplicationJoinUserRepository
	ApplicationBaseRepository
}

type applicationRepositoryImpl struct {
	DB *gorm.DB
}

var apply ApplicationRepository

func init() {
	apply = &applicationRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetApplicationRepository() ApplicationRepository {
	return apply
}

func (a *applicationRepositoryImpl) FetchApplicationsByAdmin(ctx context.Context, roomID uint64, skip int, pageSize int) (records []*model.UserJoinApplyRecord, err error) {
	tx := GetTxContext(ctx, a.DB)
	result := tx.Table("apply_records").
		Select(`apply_records."id" as id, 
				apply_records.users_id as users_id,
				users."name" as name, 
				users."email" as email`).
		Joins(`inner join users on apply_records.users_id = users."id"`).
		Where(`apply_records.room_id=?`, roomID).
		Order("id DESC").Offset(skip).Limit(pageSize).
		Scan(records)
	return records, result.Error
}

func (a *applicationRepositoryImpl) FetchApplicationsByUser(ctx context.Context, userID uint64, skip int, pageSize int) (records []*model.RoomJoinApplyRecord, err error) {
	tx := GetTxContext(ctx, a.DB)
	result := tx.Table("apply_records").
		Select(`apply_records."id" as id, 
				apply_records.room_id as room_id,
				rooms."name" as room_name, 
				rooms.admin_user_id as admin_user_id, 
				rooms.user_ids as user_ids,
				rooms.description`).
		Joins(`inner join rooms on apply_records.room_id = rooms."id"`).
		Where(`apply_records.user_id=?`, userID).
		Order("id DESC").Offset(skip).Limit(pageSize).
		Scan(&records)
	return records, result.Error
}

func (a *applicationRepositoryImpl) RoomJoinApply(ctx context.Context, roomID uint64, userID uint64) (bool, error) {
	tx := GetTxContext(ctx, a.DB)
	applyRecord := model.ApplyRecord{RoomID: roomID, UserID: userID}
	result := tx.Create(&applyRecord)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (a *applicationRepositoryImpl) RoomJoinApplyRequestDelete(ctx context.Context, roomID uint64, userID uint64) (bool, error) {
	tx := GetTxContext(ctx, a.DB)
	result := tx.Where("room_id=? and user_id=?", roomID, userID).Delete(&model.ApplyRecord{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (a *applicationRepositoryImpl) CheckApplicationExist(ctx context.Context, roomID uint64, userID uint64) (exist bool, err error) {
	tx := GetTxContext(ctx, a.DB)
	result := tx.Select("id").Where("room_id=? and user_id=?", roomID, userID).First(&model.ApplyRecord{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}
