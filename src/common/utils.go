package common

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func PQInt64ArrayToUInt64Array(arr pq.Int64Array) []uint64 {
	answer := make([]uint64, len(arr))
	for i, value := range arr {
		answer[i] = uint64(value)
	}
	return answer
}

func UInt64ArrayToPQInt64Array(arr []uint64) pq.Int64Array {
	answer := make([]int64, len(arr))
	for i, value := range arr {
		answer[i] = int64(value)
	}
	return answer
}

func TimeToUint64(t time.Time) uint64 {
	return uint64(t.UnixNano())
}

func Uint64ToTime(u uint64) time.Time {
	return time.Unix(0, int64(u))
}

func SetUUID(c *gin.Context) {
	requestID := uuid.New().String()
	c.Set("RequestID", requestID)
}

func GetUUID(c context.Context) string {
	val := c.Value("RequestID")
	uuid, _ := val.(string)
	return uuid
}
