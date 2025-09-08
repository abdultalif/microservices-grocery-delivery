package logger

import (
	"context"
	"fmt"
	"os"
	"time"
)

type FileLoggerInterface interface {
	LogOAuthActivity(ctx context.Context, userID int64, provider, action, status, message, ip, userAgent string)
}

type FileLogger struct {
	file *os.File
}

// LogOAuthActivity implements fileLoggerInterface.
func (f *FileLogger) LogOAuthActivity(ctx context.Context, userID int64, provider string, action string, status string, message string, ip string, userAgent string) {

	logLine := fmt.Sprintf(
		"%s | userID=%d | provider=%s | action=%s | status=%s | msg=%s | ip=%s | ua=%s\n",
		time.Now().Format(time.RFC3339),
		userID, provider, action, status, message, ip, userAgent,
	)
	f.file.WriteString(logLine)

}

func NewLogger(file *os.File) FileLoggerInterface {
	return &FileLogger{file: file}
}
