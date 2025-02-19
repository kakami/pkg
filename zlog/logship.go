package zlog

import (
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_defaultLogship = New(os.Stdout, zapcore.InfoLevel)
)

func Default() *Logship {
	return _defaultLogship
}

func SetWriter(w io.Writer) {
	_defaultLogship.SetWriter(w)
}

func SetDefaultLevel(l zapcore.Level) {
	_defaultLogship.SetDefaultLevel(l)
}

func LoggerWithTag(tag string) *zap.Logger {
	return _defaultLogship.LoggerWithTag(tag)
}

func SimpleLoggerWithTag(tag string) *zap.Logger {
	return _defaultLogship.SimpleLoggerWithTag(tag)
}

func SetLevel(tag string, l zapcore.Level) {
	_defaultLogship.SetLevel(tag, l)
}

func Enabled(tag string, l zapcore.Level) bool {
	return _defaultLogship.Enabled(tag, l)
}

type Logship struct {
	w     io.Writer
	lvls  sync.Map
	level zapcore.Level
}

func New(w io.Writer, level zapcore.Level) *Logship {
	return &Logship{
		w:     w,
		level: level,
	}
}

func (ls *Logship) SetWriter(w io.Writer) {
	ls.w = w
}

func (ls *Logship) SetDefaultLevel(l zapcore.Level) {
	ls.level = l
}

func (ls *Logship) LoggerWithTag(tag string) *zap.Logger {
	ptag := "[" + tag + "]"
	lvli, _ := ls.lvls.LoadOrStore(tag, zap.NewAtomicLevelAt(ls.level))
	lvl, _ := lvli.(zap.AtomicLevel)
	return DefaultLoggerWithLevel(ls.w, lvl).Named(ptag)
}

func (ls *Logship) SimpleLoggerWithTag(tag string) *zap.Logger {
	ptag := "[" + tag + "]"
	lvli, _ := ls.lvls.LoadOrStore(tag, zap.NewAtomicLevelAt(ls.level))
	lvl, _ := lvli.(zap.AtomicLevel)
	return SimpleLoggerWithLevel(ls.w, lvl).Named(ptag)
}

func (ls *Logship) SetLevel(tag string, l zapcore.Level) {
	lvli, _ := ls.lvls.LoadOrStore(tag, zap.NewAtomicLevel())
	lvl, _ := lvli.(zap.AtomicLevel)
	lvl.SetLevel(l)
}

func (ls *Logship) Enabled(tag string, l zapcore.Level) bool {
	lvli, ok := ls.lvls.Load(tag)
	if ok {
		if lvl, ok := lvli.(zap.AtomicLevel); ok {
			return lvl.Enabled(l)
		}
	}
	return false
}
