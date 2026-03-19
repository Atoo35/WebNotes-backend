package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func InitLogger() {

	// Check environment variable to determine if it's dev or prod
	env := os.Getenv("APP_ENV") // Set APP_ENV to "dev" or "prod"

	if env == "dev" || env == "local" || env == "" {
		// Development logger (console only)
		consoleLogger()
		return
	}

	ws := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "application.log",
		MaxSize:    10, //MB
		MaxBackups: 30,
		MaxAge:     90, //days
		Compress:   false,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	core := zapcore.NewCore(
		// use NewConsoleEncoder for human-readable output
		zapcore.NewJSONEncoder(encoderConfig),
		// write to stdout as well as log files
		ws,
		// Set the log level to InfoLevel, so it includes Info, Warning, Error, etc.
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	// Create the logger instance
	Logger = zap.New(core, zap.AddCaller(), zap.Development())

	// Ensure logs are flushed to output
	defer Logger.Sync()
}

func consoleLogger() {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)
	Logger = zap.New(core, zap.AddCaller(), zap.Development())
}
