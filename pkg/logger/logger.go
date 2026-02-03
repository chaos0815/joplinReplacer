package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

// Init initializes the logger
func Init(verbose bool) error {
	var err error
	if verbose {
		log, err = zap.NewDevelopment()
	} else {
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		log, err = config.Build()
	}
	return err
}

// Get returns the logger instance
func Get() *zap.Logger {
	if log == nil {
		log, _ = zap.NewProduction()
	}
	return log
}

// Sync flushes any buffered log entries
func Sync() {
	if log != nil {
		log.Sync()
	}
}
