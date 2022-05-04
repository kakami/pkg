package zlog

import (
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_w    io.Writer = os.Stdout
	_lvls sync.Map
)

func SetWriter(w io.Writer) {
	_w = w
}

func LoggerWithTag(tag string) *zap.Logger {
	ptag := "[" + tag + "]"
	lvli, _ := _lvls.LoadOrStore(tag, zap.NewAtomicLevelAt(zapcore.InfoLevel))
	lvl, _ := lvli.(zap.AtomicLevel)
	return DefaultLoggerWithLevel(_w, lvl).Named(ptag)
}

func SetLevel(tag string, l zapcore.Level) {
	lvli, _ := _lvls.LoadOrStore(tag, zap.NewAtomicLevel())
	lvl, _ := lvli.(zap.AtomicLevel)
	lvl.SetLevel(l)
}

func Enabled(tag string, l zapcore.Level) bool {
	lvli, ok := _lvls.Load(tag)
	if ok {
		if lvl, ok := lvli.(zap.AtomicLevel); ok {
			return lvl.Enabled(l)
		}
	}
	return false
}
