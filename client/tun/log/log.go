package log

import (
	"fmt"
	"sync"
)

// global Logger and SugaredLogger.
var (
	_globalMu sync.RWMutex
	_globalL  *Logger
	_globalS  *SugaredLogger
)

func init() {
	SetLogger(NewLogger(InfoLevel))
}

func NewLeveled(l Level, options ...Option) (*Logger, error) {
	switch l {
	case SilentLevel:
		return NewLogger(SilentLevel), nil
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel:
		return NewLogger(l), nil
	default:
		return nil, fmt.Errorf("invalid level: %s", l)
	}
}

// SetLogger sets the global Logger and SugaredLogger.
func SetLogger(logger *Logger) {
	_globalMu.Lock()
	defer _globalMu.Unlock()
	_globalL = logger
	_globalS = _globalL.Sugar()
	_globalE.setLogger(_globalS)
}

func logf(lvl Level, template string, args ...any) {
	_globalMu.RLock()
	s := _globalS
	_globalMu.RUnlock()
	s.Logf(lvl, template, args...)
}

func Debugf(template string, args ...any) {
	logf(DebugLevel, template, args...)
}

func Infof(template string, args ...any) {
	logf(InfoLevel, template, args...)
}

func Warnf(template string, args ...any) {
	logf(WarnLevel, template, args...)
}

func Errorf(template string, args ...any) {
	logf(ErrorLevel, template, args...)
}

func Fatalf(template string, args ...any) {
	logf(FatalLevel, template, args...)
}
