package observability

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds a JSON zap logger writing to stderr.
func NewLogger(level string) (*zap.Logger, error) {
	return newLogger(level, nil)
}

// NewFileLogger writes JSON logs to path (use from TUI so stderr stays clean).
func NewFileLogger(level, path string) (*zap.Logger, error) {
	return newLogger(level, []string{path})
}

func newLogger(level string, outputs []string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true
	if len(outputs) > 0 {
		cfg.OutputPaths = outputs
		cfg.ErrorOutputPaths = outputs
	}
	lvl := zapcore.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	return cfg.Build()
}
