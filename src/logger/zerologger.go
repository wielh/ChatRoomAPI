package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type zerologgerFactory struct {
}

func NewZeroLogger() Logger {
	return &zerologgerReciverImpl{
		logger: zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel),
	}
}

type zerologgerReciverImpl struct {
	logger zerolog.Logger
}

func (z *zerologgerReciverImpl) Debug(requestId string, checkpoint string, data any, err error) {
	z.logger.Debug().Msg(getMessage(requestId, checkpoint, data, err, 4))
}

func (z *zerologgerReciverImpl) Info(requestId string, checkpoint string, data any, err error) {
	z.logger.Info().Msg(getMessage(requestId, checkpoint, data, err, 4))
}

func (z *zerologgerReciverImpl) Warning(requestId string, checkpoint string, data any, err error) {
	z.logger.Warn().Msg(getMessage(requestId, checkpoint, data, err, 4))
}

func (z *zerologgerReciverImpl) Error(requestId string, checkpoint string, data any, err error) {
	z.logger.Error().Msg(getMessage(requestId, checkpoint, data, err, 4))
}
