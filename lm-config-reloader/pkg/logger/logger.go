package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

var levelStrings = map[string]zapcore.Level{
	"debug": zap.DebugLevel,
	"info":  zap.InfoLevel,
	"error": zap.ErrorLevel,
}

// Init intiliazes the logger
func Init(logLevel string) error {

	atom := zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()

	logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
	// nolint
	defer logger.Sync()

	if lgLevel, ok := levelStrings[strings.ToLower(logLevel)]; ok {
		atom.SetLevel(lgLevel)
	} else {
		logger.Error("unsupported log level", zap.String("logLevel", logLevel))
		logger.Info("defaulting the log level to info level")
	}
	return nil
}

// Logger returns the logger instance
func Logger() *zap.Logger {
	return logger
}
