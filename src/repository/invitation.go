package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type InvitationJoinRoomRepository interface {
	FetchInvitationsByAdmin(ctx context.Context, roomID uint64, skip int, pageSize int) (invitations []*model.UserJoinInviteRecord, err error)
}

type InvitationJoinUserRepository interface {
	FetchInvitationsByUser(ctx context.Context, userID uint64, skip int, pageSize int) (invitations []*model.RoomJoinInviteRecord, err error)
}

type InvitationBaseRepository interface {
	InviteNewUser(ctx context.Context, roomID uint64, userID uint64) error
	InviteNewUserRequestDelete(ctx context.Context, roomID uint64, userID uint64) (bool, error)
	CheckInvitationExist(ctx context.Context, roomID uint64, userID uint64) (exist bool, err error)
}

type InvitationRepository interface {
	InvitationJoinRoomRepository
	InvitationJoinUserRepository
	InvitationBaseRepository
}

type invitationRepositoryImpl struct {
	DB *gorm.DB
}

var invite InvitationRepository

func init() {
	invite = &invitationRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetInvitationRepository() InvitationRepository {
	return invite
}

func (i *invitationRepositoryImpl) FetchInvitationsByAdmin(ctx context.Context, roomID uint64, skip int, pageSize int) (invitations []*model.UserJoinInviteRecord, err error) {
	tx := GetTxContext(ctx, i.DB)
	result := tx.Table("invite_records").
		Select(`invite_records."id" as id, 
				invite_records.users_id as users_id,
				users."name" as name, 
				users."email" as email`).
		Joins(`inner join users on invite_records.users_id = users."id"`).
		Where(`invite_records.room_id=?`, roomID).
		Order("id DESC").Offset(skip).Limit(pageSize).
		Scan(invitations)
	return invitations, result.Error
}

func (i *invitationRepositoryImpl) FetchInvitationsByUser(ctx context.Context, userID uint64, skip int, pageSize int) (invitations []*model.RoomJoinInviteRecord, err error) {
	tx := GetTxContext(ctx, i.DB)
	result := tx.Table("invite_records").
		Select(`invite_records.id as id, 
				invite_records.room_id as room_id,
				rooms.name as room_name, 
				rooms.admin_user_id as admin_user_id, 
				rooms.user_ids as user_ids,
				rooms.description`).
		Joins(`inner join rooms on invite_records.room_id = rooms.id`).
		Where(`invite_records.user_id=?`, userID).
		Order("id DESC").Offset(skip).Limit(pageSize).
		Scan(&invitations)
	return invitations, result.Error
}

func (i *invitationRepositoryImpl) InviteNewUser(ctx context.Context, roomID uint64, userID uint64) error {
	tx := GetTxContext(ctx, i.DB)
	inviteRecord := model.InviteRecord{RoomID: roomID, UserID: userID}
	result := tx.Create(&inviteRecord)
	return result.Error
}

func (i *invitationRepositoryImpl) InviteNewUserRequestDelete(ctx context.Context, roomID uint64, userID uint64) (bool, error) {
	tx := GetTxContext(ctx, i.DB)
	result := tx.Where("room_id=? and user_id=?", roomID, userID).Delete(&model.InviteRecord{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (i *invitationRepositoryImpl) CheckInvitationExist(ctx context.Context, roomID uint64, userID uint64) (exist bool, err error) {
	tx := GetTxContext(ctx, i.DB)
	result := tx.Select("id").Where("room_id=? and user_id=?", roomID, userID).First(&model.InviteRecord{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}
