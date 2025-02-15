package logger

import (
	"ChatRoomAPI/src"
	"fmt"
)

func NewLogger() Logger {
	switch src.GlobalConfig.YamlConfig.Log.Level {
	case "error":
		return NewErrorLogger()
	case "warning":
		return NewWarnLogger()
	case "debug":
		return NewDebugLogger()
	default:
		return NewInfoLogger()
	}
}

func NewDebugLogger() Logger {
	return &loggerReciverImpl{
		level:    def.Debug,
		levelDef: def,
	}
}

func NewInfoLogger() Logger {
	return &loggerReciverImpl{
		level:    def.Info,
		levelDef: def,
	}
}

func NewWarnLogger() Logger {
	return &loggerReciverImpl{
		level:    def.Warning,
		levelDef: def,
	}
}

func NewErrorLogger() Logger {
	return &loggerReciverImpl{
		level:    def.Error,
		levelDef: def,
	}
}

type loggerReciverImpl struct {
	level    int
	levelDef loggerLevelDefiniton
}

type loggerLevelDefiniton struct {
	Debug   int
	Info    int
	Warning int
	Error   int
}

var def = loggerLevelDefiniton{
	Debug: 0, Info: 1, Warning: 2, Error: 3,
}

func (l *loggerReciverImpl) Debug(requestId string, checkpoint string, data any, err error) {
	if l.level <= l.levelDef.Debug {
		message := getMessage(requestId, checkpoint, data, err, 4)
		go func() {
			fmt.Printf("[%s]%s", "Debug", message)
		}()
	}
}

func (l *loggerReciverImpl) Info(requestId string, checkpoint string, data any, err error) {
	if l.level <= l.levelDef.Info {
		message := getMessage(requestId, checkpoint, data, err, 4)
		go func() {
			fmt.Printf("[%s]%s", "Info", message)
		}()
	}
}

func (l *loggerReciverImpl) Warning(requestId string, checkpoint string, data any, err error) {
	if l.level <= l.levelDef.Warning {
		message := getMessage(requestId, checkpoint, data, err, 4)
		go func() {
			fmt.Printf("[%s]%s", "Warn", message)
		}()
	}
}

func (l *loggerReciverImpl) Error(requestId string, checkpoint string, data any, err error) {
	if l.level <= l.levelDef.Error {
		message := getMessage(requestId, checkpoint, data, err, 4)
		go func() {
			fmt.Printf("[%s]%s", "Error", message)
		}()
	}
}
