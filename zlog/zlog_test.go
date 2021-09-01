package zlog_test

import (
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/kakami/pkg/util"
	"github.com/kakami/pkg/zlog"
)

var (
	_sum atomic.Int64
)

func Test_TimeFileLogWriter(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	lw := zlog.NewTimeFileLogWriter("zzzlog.log", "M", 3)
	if lw == nil {
		t.Error("eeeeerrrrrrrooooooooooorrrrrrrrrrr !!!")
		return
	}
	var g errgroup.Group

	// g.Go(func() error {
	// 	goWrite(t, lw)
	// 	return nil
	// })
	/*
		g.Go(func() error {
			goZapSugarLog(t, lw)
			return nil
		})
		g.Go(func() error {
			goZapLog(t, lw)
			return nil
		})
		g.Go(func() error {
			goZapLog(t, lw)
			return nil
		})
	*/
	g.Go(func() error {
		goForZapSugarLog(t, lw)
		return nil
	})

	t.Log(g.Wait(), _sum.Load())
	lw.Close()
}

func goWrite(t *testing.T, w io.Writer) {
	str := []byte(util.RandomString(int(rand.Int31n(200))) + "\n")
	ticker := time.NewTicker(time.Duration(rand.Int31n(300)) * time.Millisecond)
	ttl := time.Now().Unix() + 200
	for {
		select {
		case t := <-ticker.C:
			_sum.Inc()
			w.Write(str)
			if t.Unix() > ttl {
				return
			}
		}
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func goZapLog(t *testing.T, w io.Writer) {
	encoderConf := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
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
	str := util.RandomString(int(rand.Int31n(200)))
	ticker := time.NewTicker(time.Duration(rand.Int31n(300)) * time.Millisecond)
	ttl := time.Now().Unix() + 300
	for {
		select {
		case t := <-ticker.C:
			_sum.Inc()
			log.Info("zapLog", zap.String("msg", str), zap.Int64("sum", _sum.Load()))
			if t.Unix() > ttl {
				return
			}
		}
	}
}

func goZapSugarLog(t *testing.T, w io.Writer) {
	encoderConf := &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
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
	log := zap.New(core).WithOptions(zap.AddCaller()).Sugar()
	str := util.RandomString(int(rand.Int31n(200)))
	ticker := time.NewTicker(time.Duration(rand.Int31n(300)) * time.Millisecond)
	ttl := time.Now().Unix() + 300
	for {
		select {
		case t := <-ticker.C:
			_sum.Inc()
			log.Infof("sugarLog msg: %s - %d", str, _sum.Load())
			if t.Unix() > ttl {
				return
			}
		}
	}
}

func goForZapSugarLog(t *testing.T, w io.Writer) {
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
	log := zap.New(core).WithOptions(zap.AddCaller()).With(zap.String("tag", "zlog_test")).Sugar()
	str := util.RandomString(int(rand.Int31n(200)))
	for {
		_sum.Inc()
		log.Infof("for sugarLog msg: %s - %d", str, _sum.Load())
		if _sum.Load() > 20 {
			break
		}
	}
}

func Test_DefaultLoggerWithLevel(t *testing.T) {
	lvl := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	lg := zlog.DefaultLoggerWithLevel(os.Stdout, lvl).Sugar()
	lg.Debug("============")
	lg.Info("===========")
	lg.Error("==============")

	lvl.SetLevel(zapcore.InfoLevel)
	lg.Debug("============")
	lg.Info("===========")
	lg.Error("==============")

	lvl.SetLevel(zapcore.ErrorLevel)
	lg.Debug("============")
	lg.Info("===========")
	lg.Error("==============")
}