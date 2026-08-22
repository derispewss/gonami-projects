package whatsapp

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	waLog "go.mau.fi/whatsmeow/util/log"
)

func newWaLog(module string) waLog.Logger {
	debug := strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug")
	return &slogWaLogger{module: module, debug: debug}
}

type slogWaLogger struct {
	module string
	debug  bool
}

func (l *slogWaLogger) Errorf(msg string, args ...interface{}) {
	formatted := l.module + ": " + fmt.Sprintf(msg, args...)

	if strings.Contains(formatted, "received message with old counter") {
		slog.Debug(formatted)
		return
	}
	slog.Error(formatted)
}

func (l *slogWaLogger) Warnf(msg string, args ...interface{}) {
	slog.Warn(l.module + ": " + fmt.Sprintf(msg, args...))
}

func (l *slogWaLogger) Infof(msg string, args ...interface{}) {
	slog.Info(l.module + ": " + fmt.Sprintf(msg, args...))
}

func (l *slogWaLogger) Debugf(msg string, args ...interface{}) {
	if !l.debug {
		return
	}
	slog.Debug(l.module + ": " + fmt.Sprintf(msg, args...))
}

func (l *slogWaLogger) Sub(module string) waLog.Logger {
	child := *l
	child.module = l.module + "/" + module
	return &child
}
