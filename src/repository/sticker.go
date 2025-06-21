package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type StickerRepository interface {
	GetStickerSetInfo(ctx context.Context, stickerSetId uint64) (*model.StickerSet, bool, error)
	GetAllAvailableStickersInfo(ctx context.Context, UserID uint64) ([]*model.StickerSet, error)
	CheckAvailable(ctx context.Context, userID uint64, stickerSetId uint64, stickerId uint64) (*model.Sticker, bool, error)
	CheckStickerUserMappingExist(ctx context.Context, stickerSetId uint64, userId uint64) (bool, error)
	StickerSetBindingToUser(ctx context.Context, stickerSetId uint64, userId uint64) error
}

type stickerRepositoryImpl struct {
	DB *gorm.DB
}

var sticker StickerRepository

func init() {
	sticker = &stickerRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetStickerRepository() StickerRepository {
	return sticker
}

func (s *stickerRepositoryImpl) CheckAvailable(ctx context.Context, userID uint64, stickerSetId uint64, stickerId uint64) (*model.Sticker, bool, error) {
	tx := GetTxContext(ctx, s.DB)
	sticker := model.Sticker{}
	result := tx.Table("stickers As s").
		Select("s.*").
		Joins("JOIN sticker_mappings m ON s.id = m.sticker_id").
		Joins("JOIN sticker_set_user_mappings um ON um.sticker_set_id = m.sticker_set_id").
		Where("m.sticker_set_id = ?", stickerSetId).
		Where("m.sticker_id = ?", stickerId).
		Where("um.user_id = ?", userID).
		First(&sticker)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if result.Error != nil {
		return nil, false, result.Error
	}
	return &sticker, true, nil
}

func (s *stickerRepositoryImpl) GetAllAvailableStickersInfo(ctx context.Context, UserID uint64) ([]*model.StickerSet, error) {
	tx := GetTxContext(ctx, s.DB)
	var stickerSets []*model.StickerSet
	result := tx.Preload("Stickers").Table("sticker_sets s").
		Joins("JOIN sticker_set_user_mappings m ON s.id = m.sticker_set_id").
		Select("s.id", "s.name", "s.author", "s.price").
		Where("m.user_id = ?", UserID).Find(&stickerSets)

	if result.Error != nil {
		return nil, result.Error
	}
	return stickerSets, nil
}

func (s *stickerRepositoryImpl) GetStickerSetInfo(ctx context.Context, stickerSetId uint64) (*model.StickerSet, bool, error) {
	tx := GetTxContext(ctx, s.DB)
	var stickerSet model.StickerSet
	result := tx.Select("id", "name", "author", "price").Where("id = ?", stickerSetId).First(&stickerSet)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if result.Error != nil {
		return nil, false, result.Error
	}

	var stickers []*model.Sticker
	result = tx.Select("id", "name").Where("sticker_set_id = ?", stickerSetId).Find(&stickers)
	if result.Error != nil {
		return nil, false, result.Error
	}
	stickerSet.Stickers = stickers
	return &stickerSet, true, nil
}

func (s *stickerRepositoryImpl) CheckStickerUserMappingExist(ctx context.Context, stickerSetId uint64, userId uint64) (bool, error) {
	tx := GetTxContext(ctx, s.DB)
	var mapping model.StickerSetUserMapping
	result := tx.Select("id").Where("sticker_set_id = ? and user_id=?", stickerSetId, userId).First(&mapping)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	} else if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (s *stickerRepositoryImpl) StickerSetBindingToUser(ctx context.Context, stickerSetId uint64, userId uint64) error {
	tx := GetTxContext(ctx, s.DB)
	mapping := model.StickerSetUserMapping{UserId: userId, StickerSetId: stickerSetId}
	result := tx.Create(&mapping)
	return result.Error
}
