// Package ogenzap contains ogen logging utilities.
package ogenzap

import (
	"flag"
	"os"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultColorFlag returns default color flag value by checking NO_COLOR env.
//
// See https://no-color.org.
func DefaultColorFlag() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return !ok
}

// Options is options for Create.
type Options struct {
	Level     zapcore.Level
	Verbose   bool
	Color     bool
	FnOptions []zap.Option
}

// RegisterFlags registers fields of Options as flags.
func (o *Options) RegisterFlags(set *flag.FlagSet) {
	set.Var(&o.Level, "loglevel", "Zap logging level")
	set.BoolVar(&o.Verbose, "v", false, "Enable verbose logging")
	set.BoolVar(&o.Color, "color", DefaultColorFlag(), "Enable color logging")
}

// Create creates new logger for ogen.
func Create(opts Options) (*zap.Logger, error) {
	level := opts.Level
	if opts.Verbose {
		level = zap.DebugLevel
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	if !opts.Verbose {
		cfg.EncoderConfig.EncodeTime = func(time.Time, zapcore.PrimitiveArrayEncoder) {
			// Set to noop if logging is not verbose.
		}
		// Disable stacktrace and caller.
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true
	}
	if opts.Color {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := cfg.Build(opts.FnOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "create logger")
	}
	return logger, nil
}
