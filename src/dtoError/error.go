package dtoError

import (
	"github.com/gin-gonic/gin"
)

type ServiceError struct {
	StatusCode     int
	ErrorCode      int64
	InternalError  error
	ExtrenalReason string
}

func (s *ServiceError) ToJsonResponse() (statusCode int, H *gin.H) {
	statusCode = s.StatusCode
	H = &gin.H{
		"errorCode": s.ErrorCode,
		"reason":    s.ExtrenalReason,
	}
	return
}

const (
	Success             = 0
	UnKnown             = 1
	ParseFormatFailed   = 2
	ParseJsonFailed     = 3
	UserHasRegisterd    = 4
	UsernameExist       = 5
	LoginFailed         = 6
	ResetPasswordFailed = 7
	UserNotExist        = 8

	DBError          = 10000
	DBNoRowAffected  = 10001
	DBtxCommitFailed = 10002

	RoomNotExist      = 20000
	UserNotInRoom     = 20001
	NotAdminInRoom    = 20001
	UserAlreadyInRoom = 20003
	RoomNameUsed      = 20004

	UserIsInvited    = 30000
	UserIsNotInvited = 30001

	UserHasApplied = 40000
	UserNotApply   = 40001
)

type ServiceErrorWarpper interface {
	NewLoginFailedServiceError(err error) *ServiceError
	NewRessetPasswordServiceError(err error) *ServiceError
	NewParseFormatFailedServiceError(err error, msg string) *ServiceError
	NewParseJsonFailedServiceError(err error) *ServiceError // use for err=c.ShouldBindJSON(&req) only
	NewUserHasRegisterdError(username string) *ServiceError
	NewUsernameExist(username string) *ServiceError
	NewUserNotExist(Id uint64) *ServiceError

	NewDBServiceError(err error) *ServiceError
	NewDBNoAffectedServiceError() *ServiceError
	NewDBCommitServiceError(err error) *ServiceError

	NewRoomNotExistError(roomID uint64) *ServiceError
	NewUserNotInRoomError(userID uint64, roomID uint64) *ServiceError
	NewNotAdminOfRoomError(adminID uint64, roomID uint64) *ServiceError
	NewUserAlreadyInRoomError(userID uint64, roomID uint64) *ServiceError
	NewRoomNameUsedError(roomName string) *ServiceError

	NewUserIsInvitedError(userID uint64, roomID uint64) *ServiceError
	NewUserIsNotInvitedError(userID uint64, roomID uint64) *ServiceError

	NewUserApplyError(userID uint64, roomID uint64) *ServiceError
	NewUserNotApplyError(userID uint64, roomID uint64) *ServiceError
}
