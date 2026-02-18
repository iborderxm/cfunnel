package log

import (
	"fmt"
	stdlog "log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu     sync.Mutex
	logger *stdlog.Logger
	level  Level
}

type SugaredLogger struct {
	logger *Logger
}

type Option func(*Logger) *Logger

func NewLogger(level Level) *Logger {
	return &Logger{
		logger: stdlog.New(os.Stdout, "", stdlog.LstdFlags),
		level:  level,
	}
}

func (l *Logger) WithOptions(opts ...Option) *Logger {
	for _, opt := range opts {
		l = opt(l)
	}
	return l
}

func (l *Logger) Sugar() *SugaredLogger {
	return &SugaredLogger{logger: l}
}

func (l *Logger) Logf(lvl Level, template string, args ...any) {
	if lvl < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	timestamp := time.Now().Format("2006-01-02T15:04:05.999999999Z07:00")
	prefix := fmt.Sprintf("[%s] %s ", timestamp, lvl.String())
	l.logger.SetPrefix(prefix)
	l.logger.Printf(template, args...)
}

func (s *SugaredLogger) Logf(lvl Level, template string, args ...any) {
	s.logger.Logf(lvl, template, args...)
}

func (s *SugaredLogger) Debugf(template string, args ...any) {
	s.Logf(DebugLevel, template, args...)
}

func (s *SugaredLogger) Infof(template string, args ...any) {
	s.Logf(InfoLevel, template, args...)
}

func (s *SugaredLogger) Warnf(template string, args ...any) {
	s.Logf(WarnLevel, template, args...)
}

func (s *SugaredLogger) Errorf(template string, args ...any) {
	s.Logf(ErrorLevel, template, args...)
}

func (s *SugaredLogger) Fatalf(template string, args ...any) {
	s.Logf(FatalLevel, template, args...)
	os.Exit(1)
}
