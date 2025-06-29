package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ChatRoomAPI/src"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type StickerSetCacheInfo struct {
	Id       uint64
	Name     string
	Author   string
	Price    uint32
	Stickers map[uint64]*StickerCacheInfo
}

type StickerCacheInfo struct {
	Id   uint64
	Name string
}

type StickerCache interface {
	StoreStickerSetInfoByUser(ctx context.Context, userId uint64, infos map[uint64]*StickerSetCacheInfo) error
	InsertNewStickerSetInfoByUser(ctx context.Context, userId uint64, info *StickerSetCacheInfo) (bool, error)
	GetAllStickerSetInfoByUser(ctx context.Context, userId uint64) (map[uint64]*StickerSetCacheInfo, bool, error)
	CheckStickerIDValid(ctx context.Context, userId uint64, stickerSetId uint64, stickerId uint64) (bool, error)
	ClearAllStickerCacheByUser(ctx context.Context, userId uint64) error
}

type stickerCacheImpl struct {
	redisClient    *redis.Client
	keyExpiredTime time.Duration
	tracer         trace.Tracer
}

func (s *stickerCacheImpl) getUserStickerInfosKey(userId uint64) string {
	return fmt.Sprintf("sticker::user:%d", userId)
}

func (s *stickerCacheImpl) getUserStickerIdSetKey(userId uint64, stickerSetId uint64) string {
	return fmt.Sprintf("sticker::user:%d::StickerSetId:%d", userId, stickerSetId)
}

func (s *stickerCacheImpl) CheckStickerIDValid(ctx context.Context, userId uint64, stickerSetId uint64, stickerId uint64) (bool, error) {
	key := s.getUserStickerIdSetKey(userId, stickerSetId)
	isMember, err := s.redisClient.SIsMember(ctx, key, stickerId).Result()
	if err != nil {
		return false, err
	}

	s.redisClient.Expire(ctx, key, s.keyExpiredTime)
	return isMember, nil
}

func (s *stickerCacheImpl) GetAllStickerSetInfoByUser(ctx context.Context, userId uint64) (map[uint64]*StickerSetCacheInfo, bool, error) {

	key := s.getUserStickerInfosKey(userId)
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, fmt.Errorf("redis EXISTS failed: %w", err)
	}
	if exists == 0 {
		return nil, false, nil
	}

	result, err := s.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, false, fmt.Errorf("redis HGetAll failed: %w", err)
	}

	stickerSets := make(map[uint64]*StickerSetCacheInfo)
	for idStr, jsonStr := range result {
		var info StickerSetCacheInfo
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
			return nil, false, fmt.Errorf("json unmarshal failed for sticker set ID %s: %w", idStr, err)
		}

		var id uint64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			return nil, false, fmt.Errorf("invalid sticker set ID key: %s", idStr)
		}
		stickerSets[id] = &info
	}
	return stickerSets, true, nil
}

func (s *stickerCacheImpl) InsertNewStickerSetInfoByUser(ctx context.Context, userId uint64, info *StickerSetCacheInfo) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "InsertNewStickerSetInfoByUser")
	defer span.End()

	infos, keyExist, err := s.GetAllStickerSetInfoByUser(ctx, userId)
	if err != nil {
		return false, err
	} else if !keyExist {
		return false, nil
	}

	for id := range infos {
		if id == info.Id {
			return true, nil
		}
	}
	infos[info.Id] = info
	return true, s.StoreStickerSetInfoByUser(ctx, userId, infos)
}

func (s *stickerCacheImpl) StoreStickerSetInfoByUser(ctx context.Context, userId uint64, infos map[uint64]*StickerSetCacheInfo) error {
	ctx, span := s.tracer.Start(ctx, "StoreStickerSetInfoByUser")
	defer span.End()

	key := s.getUserStickerInfosKey(userId)

	hashData := make(map[string]interface{})
	for id, info := range infos {
		data, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("marshal StickerSetCacheInfo failed: %w", err)
		}
		hashData[fmt.Sprintf("%d", id)] = data
	}
	if err := s.redisClient.HSet(ctx, key, hashData).Err(); err != nil {
		return fmt.Errorf("redis HSet failed: %w", err)
	}
	if err := s.redisClient.Expire(ctx, key, s.keyExpiredTime).Err(); err != nil {
		return fmt.Errorf("set expire failed: %w", err)
	}

	for _, info := range infos {
		setKey := s.getUserStickerIdSetKey(userId, info.Id)
		ids := make([]interface{}, 0, len(info.Stickers))
		for _, sticker := range info.Stickers {
			ids = append(ids, sticker.Id)
		}
		if len(ids) > 0 {
			if err := s.redisClient.SAdd(ctx, setKey, ids...).Err(); err != nil {
				return fmt.Errorf("SAdd failed for set %s: %w", setKey, err)
			}
			if err := s.redisClient.Expire(ctx, setKey, s.keyExpiredTime).Err(); err != nil {
				return fmt.Errorf("expire failed for set %s: %w", setKey, err)
			}
		}
	}
	return nil
}

func (s *stickerCacheImpl) ClearAllStickerCacheByUser(ctx context.Context, userId uint64) error {
	ctx, span := s.tracer.Start(ctx, "ClearAllStickerCacheByUser")
	defer span.End()

	mainKey := s.getUserStickerInfosKey(userId)
	entries, err := s.redisClient.HGetAll(ctx, mainKey).Result()
	if err != nil {
		return fmt.Errorf("redis HGetAll failed: %w", err)
	}

	var keysToDelete []string
	keysToDelete = append(keysToDelete, mainKey)
	for stickerSetIdStr := range entries {
		var stickerSetId uint64
		if _, err := fmt.Sscanf(stickerSetIdStr, "%d", &stickerSetId); err != nil {
			return fmt.Errorf("invalid stickerSetId key: %s", stickerSetIdStr)
		}
		setKey := s.getUserStickerIdSetKey(userId, stickerSetId)
		keysToDelete = append(keysToDelete, setKey)
	}

	if len(keysToDelete) > 0 {
		if err := s.redisClient.Del(ctx, keysToDelete...).Err(); err != nil {
			return fmt.Errorf("redis DEL failed: %w", err)
		}
	}

	return nil
}

var sticker StickerCache

func init() {
	sticker = &stickerCacheImpl{
		keyExpiredTime: 60 * time.Minute,
		redisClient:    src.GlobalConfig.Redis,
		tracer:         otel.Tracer("stickerCache"),
	}
}

func GetStickerCache() StickerCache {
	return sticker
}
