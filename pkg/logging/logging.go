package logging

import (
	"context"
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// DefaultLevel is the default log level.
	DefaultLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// DefaultFormat is the default log format.
	DefaultFormat = FormatJSON
	// FormatConsole marks the console log format.
	FormatConsole = "console"
	// FormatJSON marks the JSON log format.
	FormatJSON = "json"
	// Logger is the default, system-wide logger.
	Logger *zap.Logger
)

//TODO: Implement service for env variables definition using flags
const (
	FLUENTD_CONNECTION_FAILED_ERROR_MSG string = "Failed to connect to Fluentd instance"
	FLUENTD_HOST                        string = "172.17.0.1"
	FLUENTD_PORT                        int    = 24224
)

func init() {
	var (
		format = "json"
		level  = DefaultLevel.String()
	)
	if v := os.Getenv("DSS_LOG_LEVEL"); v != "" {
		level = v
	}

	if v := os.Getenv("DSS_LOG_FORMAT"); v != "" {
		format = v
	}

	if err := setUpLogger(level, format, false); err != nil {
		panic(err)
	}
}

func setUpLogger(level string, format string, isFluentEnabled bool) error {
	lvl := DefaultLevel
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	l := zap.New(configureCore(isFluentEnabled))

	Logger = l

	return nil
}

func configureCore(isFluentEnabled bool) zapcore.Core {
	defaultConfig := generateEncoderDefaultConfig()
	consoleEncoder := zapcore.NewConsoleEncoder(defaultConfig)

	if isFluentEnabled {
		fluentCore := configureFluentCore(defaultConfig)
		return zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
			fluentCore,
		)
	}

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)
}
func generateEncoderDefaultConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()
	config.EncodeDuration = zapcore.StringDurationEncoder
	config.StacktraceKey = "stack"
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	return config
}

func configureFluentCore(defaultConfig zapcore.EncoderConfig) zapcore.Core {
	fluentLogger, err := fluent.New(fluent.Config{
		FluentHost: FLUENTD_HOST,
		FluentPort: FLUENTD_PORT,
	})

	if err != nil {
		panic(FLUENTD_CONNECTION_FAILED_ERROR_MSG)
	}

	fw := &FluentWriter{fluentLogger}
	fluentdEncoder := zapcore.NewJSONEncoder(defaultConfig)
	return zapcore.NewCore(fluentdEncoder, zapcore.AddSync(fw), zapcore.DebugLevel)
}

// Configure configures the default log "level" and the log "format".
func Configure(level string, format string, isFluentEnabled bool) error {
	return setUpLogger(level, format, isFluentEnabled)
}

// WithValuesFromContext augments logger with relevant fields from ctx and returns
// the resulting logger.
func WithValuesFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	// Naive implementation for now, meant to evolve over time.
	return logger
}
