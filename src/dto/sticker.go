package dto

type BuyStickerSetRequest struct {
	UserID       uint64
	StickerSetId uint64 `json:"sticker_set_id" binding:"required"`
}

type BuyStickerResponse struct{}

type CheckStickerSetAvailableRequest struct {
	UserID       uint64
	StickerSetID uint64 `form:"sticker_set_id" binding:"required"`
}

type CheckStickerSetAvailableResponse struct {
	Ok bool
}

type GetStickerSetInfoRequest struct {
	StickerSetID uint64 `form:"sticker_set_id" binding:"required"`
}

type StickerInfo struct {
	Id   uint64
	Name string
}

type StickerSetInfo struct {
	Id       uint64
	Name     string
	Author   string
	Price    uint32
	Stickers []StickerInfo
}

type GetStickerSetInfoResponse struct {
	StickerSetInfo `json:",inline"`
}

type GetAllAvailableStickersInfoRequest struct {
	UserID uint64
}

type GetAllAvailableStickersInfoResponse struct {
	StickerSetInfoList []StickerSetInfo
}
