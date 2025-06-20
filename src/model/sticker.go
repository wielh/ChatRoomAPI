package model

type StickerSet struct {
	Id         uint64 `gorm:"primaryKey"`
	Name       string
	Author     string
	Price      uint32
	FolderPath string
	Stickers   []*Sticker
	Base
}

type Sticker struct {
	Id           uint64 `gorm:"primaryKey"`
	StickerSetId uint64
	Name         string
	Filename     string
	Base
}

type StickerSetUserMapping struct {
	Id           uint64 `gorm:"primaryKey"`
	UserId       uint64
	StickerSetId uint64
	Base
}
