package dtoError

import (
	"fmt"
	"net/http"
)

var s ServiceErrorWarpper = &ServiceErrorWarpperImpl{}

type ServiceErrorWarpperImpl struct{}

func (s *ServiceErrorWarpperImpl) NewUserChargeMoneyExcessError(userID uint64, charge uint32, min uint32, max uint32) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusBadRequest,
		ErrorCode:      UserChargeMoneyExcess,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("User %d: charge account should be between %d and %d, get %d", userID, min, max, charge),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserNotChargedError(userID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusPaymentRequired,
		ErrorCode:      UserNotCharged,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("User %d has no charged record", userID),
	}
}

func (s *ServiceErrorWarpperImpl) NewStickerAlreadyBuyError(StickerSetId uint64, userID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusBadRequest,
		ErrorCode:      UserMoneyNotEnough,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("User %d has already by sticker %d", userID, StickerSetId),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserMoneyNotEnoughError(userID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusPaymentRequired,
		ErrorCode:      StickerAlreadyBuy,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("User %d does not have enough money", userID),
	}
}

func (s *ServiceErrorWarpperImpl) NewStickerSetNotExistError(StickerSetId uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusNotFound,
		ErrorCode:      StickerSetNotExist,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("StickerSet %d does not exist", StickerSetId),
	}
}

func (s *ServiceErrorWarpperImpl) NewLoginFailedServiceError(err error) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusUnauthorized,
		ErrorCode:      LoginFailed,
		InternalError:  err,
		ExtrenalReason: "LoginFailed",
	}
}

func (s *ServiceErrorWarpperImpl) NewParseFormatFailedServiceError(err error, msg string) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusBadRequest,
		ErrorCode:      ParseFormatFailed,
		InternalError:  err,
		ExtrenalReason: msg,
	}
}

func (s *ServiceErrorWarpperImpl) NewParseJsonFailedServiceError(err error) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusBadRequest,
		ErrorCode:      ParseJsonFailed,
		InternalError:  err,
		ExtrenalReason: err.Error(),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserHasRegisterdError(username string) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserAlreadyInRoom,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %s has already registered", username),
	}
}

func (s *ServiceErrorWarpperImpl) NewDBServiceError(err error) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusInternalServerError,
		ErrorCode:      DBError,
		InternalError:  err,
		ExtrenalReason: "Service Temporary Unavailable",
	}
}

func (s *ServiceErrorWarpperImpl) NewDBNoAffectedServiceError() *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusOK,
		ErrorCode:      DBNoRowAffected,
		InternalError:  nil,
		ExtrenalReason: "No Affected Data",
	}
}

func (s *ServiceErrorWarpperImpl) NewDBCommitServiceError(err error) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusInternalServerError,
		ErrorCode:      DBtxCommitFailed,
		InternalError:  err,
		ExtrenalReason: "Service Temporary Unavailable",
	}
}

func (s *ServiceErrorWarpperImpl) NewRoomNotExistError(roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusNotFound,
		ErrorCode:      RoomNotExist,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("room %d does not exist", roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserNotInRoomError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusForbidden,
		ErrorCode:      UserNotInRoom,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d is not in room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewNotAdminOfRoomError(adminID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusForbidden,
		ErrorCode:      NotAdminInRoom,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d is not admin iof room %d", adminID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserAlreadyInRoomError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserAlreadyInRoom,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d has already in room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserIsInvitedError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserIsInvited,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d has been invited into room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserIsNotInvitedError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserIsNotInvited,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d is not invited into room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserNotApplyError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserNotApply,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d does not apply into room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewUserApplyError(userID uint64, roomID uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UserHasApplied,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d has applied into room %d", userID, roomID),
	}
}

func (s *ServiceErrorWarpperImpl) NewRoomNameUsedError(roomName string) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      RoomNameUsed,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("roomName %s is used", roomName),
	}
}

func (s *ServiceErrorWarpperImpl) NewUsernameExist(username string) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusConflict,
		ErrorCode:      UsernameExist,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("username %s is used", username),
	}
}

func (s *ServiceErrorWarpperImpl) NewRessetPasswordServiceError(err error) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusUnauthorized,
		ErrorCode:      ResetPasswordFailed,
		InternalError:  err,
		ExtrenalReason: "ResetPasswordFailed",
	}
}

func (s *ServiceErrorWarpperImpl) NewUserNotExist(Id uint64) *ServiceError {
	return &ServiceError{
		StatusCode:     http.StatusUnauthorized,
		ErrorCode:      ResetPasswordFailed,
		InternalError:  nil,
		ExtrenalReason: fmt.Sprintf("user %d does not exist", Id),
	}
}

func GetServiceErrorWarpper() ServiceErrorWarpper {
	return s
}
