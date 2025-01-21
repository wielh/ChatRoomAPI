package logger

import (
	"fmt"
	"runtime"
	"time"
)

type Logger interface {
	SetLevel(level int)
	Debug(requestId string, checkpoint string, data map[string]any, err error)
	Info(requestId string, checkpoint string, data map[string]any, err error)
	Warning(requestId string, checkpoint string, data map[string]any, err error)
	Error(requestId string, checkpoint string, data map[string]any, err error)
}

type loggerLevelDefiniton struct {
	Debug   int
	Info    int
	Warning int
	Error   int
}

func NewLogger(level int) Logger {
	return &loggerReciverImpl{
		levelDef: loggerLevelDefiniton{
			Debug: 0, Info: 1, Warning: 2, Error: 3,
		},
	}
}

type loggerReciverImpl struct {
	level    int
	levelDef loggerLevelDefiniton
}

func (l *loggerReciverImpl) Debug(requestId string, checkpoint string, data map[string]any, err error) {
	if l.level >= l.levelDef.Debug {
		l.execute("Debug", requestId, checkpoint, data, err)
	}
}

func (l *loggerReciverImpl) Info(requestId string, checkpoint string, data map[string]any, err error) {
	if l.level >= l.levelDef.Info {
		l.execute("Info", requestId, checkpoint, data, err)
	}
}

func (l *loggerReciverImpl) Warning(requestId string, checkpoint string, data map[string]any, err error) {
	if l.level >= l.levelDef.Warning {
		l.execute("Warning", requestId, checkpoint, data, err)
	}
}

func (l *loggerReciverImpl) Error(requestId string, checkpoint string, data map[string]any, err error) {
	if l.level >= l.levelDef.Error {
		l.execute("Error", requestId, checkpoint, data, err)
	}
}

func (l *loggerReciverImpl) SetLevel(level int) {
	if level < l.levelDef.Debug {
		l.level = l.levelDef.Debug
	} else if level > l.levelDef.Error {
		l.level = l.levelDef.Error
	}
	l.level = level
}

func (l *loggerReciverImpl) getCaller(i int) string {
	pc := make([]uintptr, 5)
	n := runtime.Callers(0, pc)
	if i > n {
		i = n
	}
	fn := runtime.FuncForPC(pc[i])
	return fn.Name()
}

func (l *loggerReciverImpl) execute(level string, requestId string, checkpoint string, data map[string]any, err error) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	} else {
		errStr = ""
	}
	message := fmt.Sprintf("[%s][%s] requestId: %s, caller: %s, checkpoint: %s, error: %s, data:%+v",
		level, time.Now().Format("2006-01-02 15:04:05"), requestId, l.getCaller(1), checkpoint, errStr, data)
	fmt.Println(message)
}
