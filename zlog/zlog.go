package zlog

import (
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LogBufferLength = 1024
	LogWithBlocking = true
)

// SetLogBufferLength (default is 1024)
// This should be invoked before create logWriter
func SetLogBufferLength(bufferLen int) {
	LogBufferLength = bufferLen
}

// SetLogWithBlocking (default is true)
// This should be invoked before create logWriter
func SetLogWithBlocking(isBlocking bool) {
	LogWithBlocking = isBlocking
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func DefaultLogger(w io.Writer) *zap.Logger {
	encoderConf := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(*encoderConf),
		zapcore.AddSync(w),
		atomicLevel,
	)
	log := zap.New(core).WithOptions(zap.AddCaller())

	return log
}

func DefaultLoggerWithLevel(w io.Writer, lvl zap.AtomicLevel) *zap.Logger {
	encoderConf := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(*encoderConf),
		zapcore.AddSync(w),
		lvl,
	)
	log := zap.New(core).WithOptions(zap.AddCaller())

	return log
}

func SimpleLoggerWithLevel(w io.Writer, lvl zap.AtomicLevel) *zap.Logger {
	encoderConf := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(*encoderConf),
		zapcore.AddSync(w),
		lvl,
	)
	log := zap.New(core).WithOptions(zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))

	return log
}

func NewLogger(w io.Writer, cfg *zapcore.EncoderConfig, lvl zap.AtomicLevel) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(*cfg),
		zapcore.AddSync(w),
		lvl,
	)
	log := zap.New(core).WithOptions(zap.AddCaller())

	return log
}
