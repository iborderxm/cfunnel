package log

type Level int

// Levels are aliases for Level.
const (
	DebugLevel   Level = 0
	InfoLevel    Level = 1
	WarnLevel    Level = 2
	ErrorLevel   Level = 3
	DPanicLevel  Level = 4
	PanicLevel   Level = 5
	FatalLevel   Level = 6
	InvalidLevel Level = -1
	SilentLevel  Level = 7
)

// ParseLevel is a thin wrapper for ParseLevel.
func ParseLevel(text string) (Level, error) {
	switch text {
	case "silent", "SILENT":
		return SilentLevel, nil
	case "debug", "DEBUG":
		return DebugLevel, nil
	case "info", "INFO":
		return InfoLevel, nil
	case "warn", "WARNING":
		return WarnLevel, nil
	case "error", "ERROR":
		return ErrorLevel, nil
	default:
		return InvalidLevel, nil
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case DPanicLevel:
		return "DPANIC"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	case SilentLevel:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}
