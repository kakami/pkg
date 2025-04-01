package zlog

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileLogWriter struct {
	LogCloser

	rec chan *logMsg

	// The opened file
	filename     string
	baseFilename string // abs path
	file         *os.File
}

func NewFileLogWriter(fname string) *FileLogWriter {
	w := &FileLogWriter{
		rec:      make(chan *logMsg, LogBufferLength),
		filename: fname,
	}

	w.LogCloserInit()

	if path, err := filepath.Abs(fname); err != nil {
		fmt.Fprintf(os.Stderr, "NewFileLogWriter(%s): %s\n", w.filename, err.Error())
		return nil
	} else {
		w.baseFilename = path
	}

	if err := w.intRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "NewFileLogWriter(%s): %s\n", w.filename, err.Error())
		return nil
	}

	go func() {
		defer func() {
			if w.file != nil {
				w.file.Close()
			}
		}()

		for rec := range w.rec {
			if w.EndNotify(rec) {
				return
			}

			w.file.WriteString(rec.str)
		}
	}()

	return w
}

func (w *FileLogWriter) Write(msg []byte) (int, error) {
	if !LogWithBlocking {
		if len(w.rec) >= LogBufferLength {
			return 0, ErrChannelOverflowed
		}
	}
	w.rec <- &logMsg{str: string(msg)}
	return len(msg), nil
}

func (w *FileLogWriter) Close() {
	w.WaitForEnd(w.rec)
	close(w.rec)
}

// If this is called in a threaded context, it MUST be synchronized
func (w *FileLogWriter) intRotate() error {
	// Close any log file that may be open
	if w.file != nil {
		w.file.Close()
	}

	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd

	return nil
}
