package zlog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jehiah/go-strftime"
)

const (
	MIDNIGHT = 24 * 60 * 60 /* number of seconds in a day */
	NEXTHOUR = 60 * 60      /* number of seconds in a hour */
)

var (
	ErrChannelOverflowed = errors.New("log channel overflowed")
)

type logMsg struct {
	str string
}

// WhenIsValid checks whether value of when is valid
func WhenIsValid(when string) bool {
	switch strings.ToUpper(when) {
	case "MIDNIGHT", "NEXTHOUR", "M", "H", "D":
		return true
	default:
		return false
	}
}

type TimeFileLogWriter struct {
	LogCloser

	rec chan *logMsg

	// The opened file
	filename     string
	baseFilename string // abs path
	file         *os.File

	when        string // 'D', 'H', 'M', "MIDNIGHT", "NEXTHOUR"
	backupCount int    // If backupCount is > 0, when rollover is done,

	interval   int64
	suffix     string         // suffix of log file
	fileFilter *regexp.Regexp // for removing old log files

	rolloverAt int64 // time.Unix()
}

func NewTimeFileLogWriter(fname, when string, backupCount int) *TimeFileLogWriter {
	if !WhenIsValid(when) {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%s): invalid value of when:%s\n", fname, when)
		return nil
	}

	when = strings.ToUpper(when)
	w := &TimeFileLogWriter{
		rec:         make(chan *logMsg, LogBufferLength),
		filename:    fname,
		when:        when,
		backupCount: backupCount,
	}

	w.LogCloserInit()

	if path, err := filepath.Abs(fname); err != nil {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%s): %s\n", w.filename, err.Error())
		return nil
	} else {
		w.baseFilename = path
	}

	w.prepare()

	if err := w.intRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%s): %s\n", w.filename, err.Error())
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

			if w.shouldRollover() {
				if err := w.intRotate(); err != nil {
					fmt.Fprintf(os.Stderr, "NewTimeFileLogWrite(%s): %s\n", w.filename, err.Error())
					return
				}
			}

			w.file.WriteString(rec.str)
		}
	}()

	return w
}

func (w *TimeFileLogWriter) Write(msg []byte) (int, error) {
	if !LogWithBlocking {
		if len(w.rec) >= LogBufferLength {
			return 0, ErrChannelOverflowed
		}
	}
	w.rec <- &logMsg{str: string(msg)}
	return len(msg), nil
}

func (w *TimeFileLogWriter) Close() {
	w.WaitForEnd(w.rec)
	close(w.rec)
}

func (w *TimeFileLogWriter) computeRollover(currTime time.Time) int64 {
	var result int64

	if w.when == "MIDNIGHT" {
		t := currTime.Local()
		/* r is the number of seconds left between now and midnight */
		r := MIDNIGHT - ((t.Hour()*60+t.Minute())*60 + t.Second())
		result = currTime.Unix() + int64(r)
	} else if w.when == "NEXTHOUR" {
		t := currTime.Local()
		/* r is the number of seconds left between now and the next hour */
		r := NEXTHOUR - (t.Minute()*60 + t.Second())
		result = currTime.Unix() + int64(r)
	} else {
		result = currTime.Unix() + w.interval
	}
	return result
}

// prepare prepares according to "when"
func (w *TimeFileLogWriter) prepare() {
	var regRule string

	switch w.when {
	case "M":
		w.interval = 60
		w.suffix = "%Y-%m-%d_%H-%M"
		regRule = `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}$`
	case "H", "NEXTHOUR":
		w.interval = 60 * 60
		w.suffix = "%Y-%m-%d_%H"
		regRule = `^\d{4}-\d{2}-\d{2}_\d{2}$`
	case "D", "MIDNIGHT":
		w.interval = 60 * 60 * 24
		w.suffix = "%Y-%m-%d"
		regRule = `^\d{4}-\d{2}-\d{2}$`
	default:
		// default is "D"
		w.interval = 60 * 60 * 24
		w.suffix = "%Y-%m-%d"
		regRule = `^\d{4}-\d{2}-\d{2}$`
	}
	w.fileFilter = regexp.MustCompile(regRule)

	fInfo, err := os.Stat(w.filename)

	var t time.Time
	if err == nil {
		t = fInfo.ModTime()
	} else {
		t = time.Now()
	}

	w.rolloverAt = w.computeRollover(t)
}

func (w *TimeFileLogWriter) shouldRollover() bool {
	t := time.Now().Unix()

	if t >= w.rolloverAt {
		return true
	} else {
		return false
	}
}

// If this is called in a threaded context, it MUST be synchronized
func (w *TimeFileLogWriter) intRotate() error {
	// Close any log file that may be open
	if w.file != nil {
		w.file.Close()
	}

	if w.shouldRollover() {
		// rename file to backup name
		if err := w.moveToBackup(); err != nil {
			return err
		}
	}

	// remove files, according to backupCount
	if w.backupCount > 0 {
		for _, fileName := range w.getFilesToDelete() {
			os.Remove(fileName)
		}
	}

	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd

	// adjust rolloverAt
	w.adjustRolloverAt()

	return nil
}

func (w *TimeFileLogWriter) adjustRolloverAt() {
	currTime := time.Now()
	newRolloverAt := w.computeRollover(currTime)

	for newRolloverAt <= currTime.Unix() {
		newRolloverAt = newRolloverAt + w.interval
	}

	w.rolloverAt = newRolloverAt
}

// getFilesToDelete determines the files to delete when rolling over
func (w *TimeFileLogWriter) getFilesToDelete() []string {
	dirName := filepath.Dir(w.baseFilename)
	baseName := filepath.Base(w.baseFilename)

	var result []string

	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return result
	}

	prefix := baseName + "."
	plen := len(prefix)

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if len(fileName) >= plen {
			if fileName[:plen] == prefix {
				suffix := fileName[plen:]
				if w.fileFilter.MatchString(suffix) {
					result = append(result, filepath.Join(dirName, fileName))
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(result))

	if len(result) < w.backupCount {
		result = result[0:0]
	} else {
		result = result[:len(result)-w.backupCount]
	}
	return result
}

// moveToBackup renames file to backup name
func (w *TimeFileLogWriter) moveToBackup() error {
	_, err := os.Lstat(w.filename)
	if err == nil { // file exists
		// get the time that this sequence started at and make it a TimeTuple
		t := time.Unix(w.rolloverAt-w.interval, 0).Local()
		fname := w.baseFilename + "." + strftime.Format(w.suffix, t)

		// remove the file with fname if exist
		if _, err := os.Stat(fname); err == nil {
			err = os.Remove(fname)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}

		// Rename the file to its newfound home
		err = os.Rename(w.baseFilename, fname)
		if err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}
	}
	return nil
}

/****** LogCloser ******/
type LogCloser struct {
	IsEnd chan bool
}

func (lc *LogCloser) LogCloserInit() {
	lc.IsEnd = make(chan bool)
}

// notyfy the logger log to end
func (lc *LogCloser) EndNotify(lr *logMsg) bool {
	if lr == nil && lc.IsEnd != nil {
		lc.IsEnd <- true
		return true
	}
	return false
}

// add nil to end of res and wait that EndNotify is call
func (lc *LogCloser) WaitForEnd(rec chan *logMsg) {
	rec <- nil
	if lc.IsEnd != nil {
		<-lc.IsEnd
	}
}
