package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/common"
	"ChatRoomAPI/src/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, adminUserID uint64, roomName string, description string) (room *model.Room, roomNameUsed bool, err error)
	RoomExist(ctx context.Context, roomID uint64) (bool, error)
	ReadRoomInfo(ctx context.Context, roomID uint64) (*model.Room, error)
	GetAvailbleRooms(ctx context.Context, userID uint64) ([]*model.Room, error)
	DeleteRoom(ctx context.Context, roomID uint64, adminUserID uint64) (ok bool, err error)

	AddUser(ctx context.Context, roomID uint64, userID uint64) (ok bool, err error)
	AdminChange(ctx context.Context, roomID uint64, adminUserID uint64, userID uint64) (ok bool, err error)
	CheckUserInRoom(ctx context.Context, roomID uint64, userID uint64) (isUser bool, err error)
	CheckAdminUserInRoom(ctx context.Context, roomID uint64, adminUserID uint64) (isAdmin bool, err error)
	DeleteUser(ctx context.Context, roomID uint64, adminUserID uint64, userID uint64) (ok bool, err error)
}

type roomRepositoryImpl struct {
	DB *gorm.DB
}

var room RoomRepository

func init() {
	room = &roomRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetRoomRepository() RoomRepository {
	return room
}

func (r *roomRepositoryImpl) CreateRoom(ctx context.Context, adminUserID uint64, roomName string, description string) (*model.Room, bool, error) {
	tx := GetTxContext(ctx, r.DB)
	room := model.Room{
		AdminUserID: adminUserID,
		UserIDs:     common.UInt64ArrayToPQInt64Array([]uint64{adminUserID}),
		Name:        roomName,
		Description: description,
	}
	result := tx.Where("name=?", roomName).FirstOrCreate(&room)
	if result.Error != nil {
		return nil, false, result.Error
	}
	return &room, result.RowsAffected > 0, nil
}

func (r *roomRepositoryImpl) RoomExist(ctx context.Context, roomID uint64) (bool, error) {
	tx := GetTxContext(ctx, r.DB)
	result := tx.Select("id").Where("id=?", roomID).First(&model.Room{})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (r *roomRepositoryImpl) ReadRoomInfo(ctx context.Context, roomID uint64) (*model.Room, error) {
	tx := GetTxContext(ctx, r.DB)
	roomInfo := model.Room{}
	result := tx.Select("id", "name", "admin_user_id", "user_ids", "description").Where("id=?", roomID).First(&roomInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &roomInfo, nil
}

func (r *roomRepositoryImpl) GetAvailbleRooms(ctx context.Context, userID uint64) ([]*model.Room, error) {
	tx := GetTxContext(ctx, r.DB)
	roomsInfo := []*model.Room{}
	result := tx.Select("id", "name", "admin_user_id", "user_ids", "description").Where("? = ANY (user_ids)", userID).Find(&roomsInfo)
	return roomsInfo, result.Error
}

func (r *roomRepositoryImpl) DeleteRoom(ctx context.Context, roomID uint64, adminUserID uint64) (ok bool, err error) {
	tx := GetTxContext(ctx, r.DB)
	result := tx.Where("id=? and admin_user_id=?", roomID, adminUserID).Delete(&model.Room{})
	if result.Error != nil {
		return false, err
	} else {
		return result.RowsAffected > 0, err
	}
}

func (r *roomRepositoryImpl) AddUser(ctx context.Context, roomID uint64, userID uint64) (ok bool, err error) {
	tx := GetTxContext(ctx, r.DB)
	var roomInfo model.Room
	result := tx.Where("id=?", roomID).First(&roomInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	result = tx.Model(&model.Room{}).Where("id=?", roomID).Update("user_ids", gorm.Expr("array_append(user_ids, ?)", int64(userID)))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *roomRepositoryImpl) AdminChange(ctx context.Context, roomID uint64, adminUserID uint64, userID uint64) (ok bool, err error) {
	tx := GetTxContext(ctx, r.DB)
	result := tx.Where("id=? and admin_user_id=?", roomID, adminUserID).Update("admin_user_id", userID)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *roomRepositoryImpl) CheckAdminUserInRoom(ctx context.Context, roomID uint64, adminUserID uint64) (isAdmin bool, err error) {
	tx := GetTxContext(ctx, r.DB)
	result := tx.Select("id").Where("id=? and admin_user_id=?", roomID, adminUserID).First(&model.Room{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (r *roomRepositoryImpl) CheckUserInRoom(ctx context.Context, roomID uint64, userID uint64) (isUser bool, err error) {
	tx := GetTxContext(ctx, r.DB)
	var roomInfo model.Room
	result := tx.Select("id").Where("id=? and ?=ANY (user_ids)", roomID, userID).First(&roomInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	for _, u := range roomInfo.UserIDs {
		if uint64(u) == userID {
			return true, nil
		}
	}
	return true, nil
}

func (r *roomRepositoryImpl) DeleteUser(ctx context.Context, roomID uint64, adminUserID uint64, userID uint64) (bool, error) {
	tx := GetTxContext(ctx, r.DB)
	var roomInfo model.Room
	result := tx.Where("id=?", roomID).First(&roomInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	newUsers := []int64{}
	for _, u := range roomInfo.UserIDs {
		if u == int64(userID) {
			continue
		}
		newUsers = append(newUsers, u)
	}
	result = tx.Where("room=?", roomID).Update("user_ids", newUsers)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
