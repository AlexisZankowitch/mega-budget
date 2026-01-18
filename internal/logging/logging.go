package logging

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(parseLevel(os.Getenv("LOG_LEVEL")))
	return cfg.Build()
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
