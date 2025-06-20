package model

type Wallet struct {
	Id     uint64 `gorm:"primaryKey"`
	UserID uint64
	Money  uint32
	Base
}

type WalletLog struct {
	Id     uint64 `gorm:"primaryKey"`
	UserID uint64
	Money  uint32
	Type   int32 // 0: charge 1:cost
	Detail string
	Base
}
