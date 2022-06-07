// Package ogenzap contains ogen logging utilities.
package ogenzap

import (
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Create creates new logger for ogen.
func Create(level zapcore.Level, verbose bool) (*zap.Logger, error) {
	if verbose {
		level = zap.DebugLevel
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	if !verbose {
		cfg.EncoderConfig.EncodeTime = func(time.Time, zapcore.PrimitiveArrayEncoder) {
			// Set to noop if logging is not verbose.
		}
		// Disable stacktrace and caller.
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "create logger")
	}
	return logger, nil
}
