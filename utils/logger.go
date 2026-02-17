package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel enumerates severity tiers.
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

func (l LogLevel) String() string {
	if int(l) < len(levelNames) {
		return levelNames[l]
	}
	return "UNKNOWN"
}

// Logger is a concurrency-safe, levelled logger used across the pipeline.
type Logger struct {
	mu    sync.Mutex
	level LogLevel
	inner *log.Logger
	file  *os.File
}

var (
	globalLogger *Logger
	logOnce      sync.Once
)

// InitLogger creates the singleton logger. Call once at startup.
func InitLogger(minLevel LogLevel, logFilePath string) *Logger {
	logOnce.Do(func() {
		var writers []io.Writer
		writers = append(writers, os.Stdout)

		var f *os.File
		if logFilePath != "" {
			var err error
			f, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				writers = append(writers, f)
			} else {
				log.Printf("[WARN] could not open log file %s: %v\n", logFilePath, err)
			}
		}

		mw := io.MultiWriter(writers...)
		globalLogger = &Logger{
			level: minLevel,
			inner: log.New(mw, "", 0),
			file:  f,
		}
	})
	return globalLogger
}

// L returns the global logger (panics if InitLogger has not been called).
func L() *Logger {
	if globalLogger == nil {
		// fallback: initialise a stdout-only logger at DEBUG
		return InitLogger(DEBUG, "")
	}
	return globalLogger
}

// Close flushes and closes the log file, if any.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		_ = l.file.Close()
	}
}

func (l *Logger) log(lvl LogLevel, format string, args ...any) {
	if lvl < l.level {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	l.inner.Printf("[%s] %s  %s", lvl, ts, msg)
	l.mu.Unlock()

	if lvl == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Debug(f string, a ...any) { l.log(DEBUG, f, a...) }
func (l *Logger) Info(f string, a ...any)  { l.log(INFO, f, a...) }
func (l *Logger) Warn(f string, a ...any)  { l.log(WARN, f, a...) }
func (l *Logger) Error(f string, a ...any) { l.log(ERROR, f, a...) }
func (l *Logger) Fatal(f string, a ...any) { l.log(FATAL, f, a...) }
