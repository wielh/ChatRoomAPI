package cache

import (
	"context"
)

type StickerSetCacheInfo struct {
	Id         uint64
	Name       string
	Author     string
	Price      uint32
	FolderPath string
	Stickers   []*StickerCacheInfo
}

type StickerCacheInfo struct {
	Id           uint64
	StickerSetId uint64
	Name         string
	Filename     string
}

type StickerCache interface {
	StoreStickerSetInfoByUser(ctx context.Context, userId uint64, infos []*StickerSetCacheInfo) error
	InsertNewStickerSetInfoByUser(ctx context.Context, userId uint64, info *StickerSetCacheInfo) error
	GetAllStickerSetInfoByUser(ctx context.Context, userId uint64) ([]*StickerSetCacheInfo, error)
	CheckStickerIDValid(ctx context.Context, userId uint64, stickerSetId uint64, stickerId uint64) (bool, error)
}

type stickerCacheImpl struct {
}

// CheckStickerIDValid implements StickerCache.
func (s *stickerCacheImpl) CheckStickerIDValid(ctx context.Context, userId uint64, stickerSetId uint64, stickerId uint64) (bool, error) {
	panic("unimplemented")
}

// GetAllStickerSetInfoByUser implements StickerCache.
func (s *stickerCacheImpl) GetAllStickerSetInfoByUser(ctx context.Context, userId uint64) ([]*StickerSetCacheInfo, error) {
	panic("unimplemented")
}

// InsertNewStickerSetInfoByUser implements StickerCache.
func (s *stickerCacheImpl) InsertNewStickerSetInfoByUser(ctx context.Context, userId uint64, info *StickerSetCacheInfo) error {
	panic("unimplemented")
}

// StoreStickerSetInfoByUser implements StickerCache.
func (s *stickerCacheImpl) StoreStickerSetInfoByUser(ctx context.Context, userId uint64, infos []*StickerSetCacheInfo) error {
	panic("unimplemented")
}

var sticker StickerCache

func init() {
	sticker = &stickerCacheImpl{}
}

func GetStickerCache() StickerCache {
	return sticker
}
