package repository

import (
	"ChatRoomAPI/src"
	"ChatRoomAPI/src/model"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type WalletRepository interface {
	GetState(ctx context.Context, userID uint64) (*model.Wallet, bool, error)
	WalletInit(ctx context.Context, userID uint64) error
	Charge(ctx context.Context, userID uint64, money uint32) error
	Cost(ctx context.Context, userID uint64, money uint32) (*model.Wallet, bool, error)
	WriteLog(ctx context.Context, userID uint64, Type int32, money uint32, detail string) (*model.WalletLog, error)
	GetLog(ctx context.Context, userID uint64, timeCursor time.Time, resultMaxSize int32) ([]*model.WalletLog, time.Time, error)
}

type walletRepositoryImpl struct {
	DB *gorm.DB
}

var wallet WalletRepository

func init() {
	wallet = &walletRepositoryImpl{DB: src.GlobalConfig.DB}
}

func GetWalletRepository() WalletRepository {
	return wallet
}

func (w *walletRepositoryImpl) GetState(ctx context.Context, userID uint64) (*model.Wallet, bool, error) {
	tx := GetTxContext(ctx, w.DB)
	var wallet model.Wallet
	result := tx.Select("money").Where("user_id= ?", userID).First(&wallet)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if result.Error != nil {
		return nil, false, result.Error
	}
	return &wallet, true, nil
}

func (w *walletRepositoryImpl) WalletInit(ctx context.Context, userID uint64) error {
	tx := GetTxContext(ctx, w.DB)
	wallet := model.Wallet{
		UserID: userID,
		Money:  0,
	}
	result := tx.FirstOrCreate(&wallet, model.Wallet{UserID: userID})
	return result.Error
}

func (w *walletRepositoryImpl) Charge(ctx context.Context, userID uint64, money uint32) error {
	tx := GetTxContext(ctx, w.DB)
	result := tx.Model(&model.Wallet{}).Where("user_id=?", userID).Update("money", gorm.Expr("money + ?", money))
	return result.Error
}

func (w *walletRepositoryImpl) Cost(ctx context.Context, userID uint64, money uint32) (*model.Wallet, bool, error) {
	tx := GetTxContext(ctx, w.DB)
	result := tx.Model(&model.Wallet{}).Where("user_id=? and money>=?", userID, money).Update("money", gorm.Expr("money - ?", money))
	if result.Error != nil {
		return nil, false, result.Error
	} else if result.RowsAffected == 0 {
		return nil, false, nil
	}

	var updated model.Wallet
	if err := tx.Where("user_id = ?", userID).First(&updated).Error; err != nil {
		return nil, false, err
	}
	return &updated, true, nil
}

func (w *walletRepositoryImpl) GetLog(ctx context.Context, userID uint64, timeCursor time.Time, resultMaxSize int32) (
	records []*model.WalletLog, nextTimeCursor time.Time, err error) {
	if resultMaxSize <= 0 {
		return records, nextTimeCursor, nil
	}
	tx := GetTxContext(ctx, w.DB)
	result := tx.Select("").Where("user_id=? and create_time>=?", userID, timeCursor).Order("create_time ASC").Limit(int(resultMaxSize)).Find(&records)
	if result.Error != nil {
		err = result.Error
		return
	} else if len(records) == 0 {
		nextTimeCursor = timeCursor
		return
	}
	nextTimeCursor = records[len(records)-1].CreatedAt
	return
}

func (w *walletRepositoryImpl) WriteLog(ctx context.Context, userID uint64, Type int32, money uint32, detail string) (*model.WalletLog, error) {
	tx := GetTxContext(ctx, w.DB)
	log := model.WalletLog{UserID: userID, Type: Type, Money: money, Detail: detail}
	result := tx.Create(&log)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}
