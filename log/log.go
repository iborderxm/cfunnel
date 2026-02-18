package log

import (
	"fmt"
	stdlog "log"
	"os"
	"sync"
	"time"
)

var (
	logCh = make(chan Event)
	_     = NewObservable[Event](logCh)
	level  = INFO
	logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags)
	mu     sync.Mutex
)

func init() {
	stdlog.SetFlags(stdlog.LstdFlags)
	stdlog.SetOutput(os.Stdout)
}

type Event struct {
	LogLevel LogLevel
	Payload  string
}

func (e *Event) Type() string {
	return e.LogLevel.String()
}

func Infoln(format string, v ...any) {
	event := newLog(INFO, format, v...)
	logCh <- event
	print(event)
}

func Warnln(format string, v ...any) {
	event := newLog(WARNING, format, v...)
	logCh <- event
	print(event)
}

func Errorln(format string, v ...any) {
	event := newLog(ERROR, format, v...)
	logCh <- event
	print(event)
}

func Debugln(format string, v ...any) {
	event := newLog(DEBUG, format, v...)
	logCh <- event
	print(event)
}

func Fatalln(format string, v ...any) {
	mu.Lock()
	defer mu.Unlock()
	logger.Fatalf(format, v...)
}

func print(data Event) {
	if data.LogLevel < level {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.999999999Z07:00")
	prefix := fmt.Sprintf("[%s] %s ", timestamp, data.LogLevel.String())
	logger.SetPrefix(prefix)
	logger.Println(data.Payload)
}

func newLog(logLevel LogLevel, format string, v ...any) Event {
	return Event{
		LogLevel: logLevel,
		Payload:  fmt.Sprintf(format, v...),
	}
}
