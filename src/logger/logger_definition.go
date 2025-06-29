package logger

import (
	"fmt"
	"runtime"
	"time"
)

type LoggerFactory interface {
	NewLogger() Logger
	NewDebugLogger() Logger
	NewInfoLogger() Logger
	NewWarnLogger() Logger
	NewErrorLogger() Logger
}

type Logger interface {
	Debug(requestId string, checkpoint string, data any, err error)
	Info(requestId string, checkpoint string, data any, err error)
	Warning(requestId string, checkpoint string, data any, err error)
	Error(requestId string, checkpoint string, data any, err error)
}

func getCaller(i int) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(0, pc)
	if i > n {
		i = n
	}
	fn := runtime.FuncForPC(pc[i])
	return fn.Name()
}

func getMessage(requestId string, checkpoint string, data any, err error, callerIndex int) string {
	caller := getCaller(callerIndex)
	time := time.Now().Format("2006-01-02 15:04:05.000")
	var errStr string
	if err != nil {
		errStr = err.Error()
	} else {
		errStr = ""
	}
	message := fmt.Sprintf("[%s] requestId:{%s}, caller:{%s}, checkpoint:{%s}, error:{%s}, data:{%+v}\n", time, requestId, caller, checkpoint, errStr, data)
	return message
}
