package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota + 1
	InfoLevel
	WarnLevel
	ErrorLevel
	NoLevel
	TraceLevel
)

// Logger wraps zerolog.Logger but keeps the same API
type Logger struct {
	Level   LogLevel
	Prefix  string
	nocolor bool
	zlog    zerolog.Logger
}

// NoColor disables colorized output.
func (l *Logger) NoColor(nocolor ...bool) *Logger {
	if len(nocolor) > 0 {
		l.nocolor = nocolor[0]
	} else {
		l.nocolor = true
	}
	l.configureLogger()
	return l
}

// Color returns true if colorized output is enabled.
func (l *Logger) Color() bool {
	return !l.nocolor
}

func (l *Logger) SetPrefix(prefix string) *Logger {
	l.Prefix = prefix
	l.configureLogger()
	return l
}

func (l *Logger) Lev() LogLevel {
	return l.Level
}

func (l *Logger) SetLevel(level LogLevel) *Logger {
	l.Level = level
	l.configureLogger()
	return l
}

func (l *Logger) Error(v ...any) {
	if l.Level <= ErrorLevel {
		l.zlog.Error().Str("prefix", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Warn(v ...any) {
	if l.Level <= WarnLevel {
		l.zlog.Warn().Str("prefix", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Info(v ...any) {
	if l.Level <= InfoLevel {
		l.zlog.Info().Str("prefix", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Debug(v ...any) {
	if l.Level <= DebugLevel {
		l.zlog.Debug().Str("prefix", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Trace(v ...any) {
	if l.Level <= TraceLevel {
		l.zlog.Trace().Str("prefix", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Panic(v ...any) {
	stack := make([]byte, 2048)
	runtime.Stack(stack, false)
	l.zlog.Panic().
		Str("prefix", l.Prefix).
		Str("stack", string(stack)).
		Msg(getVariable(v...))
}

func NewLogger(prefix string) *Logger {
	l := &Logger{
		Prefix: prefix,
		Level:  InfoLevel,
	}
	l.configureLogger()
	return l
}

func (l *Logger) configureLogger() {
	var output zerolog.ConsoleWriter
	output.Out = os.Stdout
	output.TimeFormat = time.RFC3339
	output.NoColor = l.nocolor

	z := zerolog.New(output).With().Timestamp().Logger()

	// Map our custom LogLevel to zerolog Level
	switch l.Level {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case InfoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WarnLevel:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ErrorLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case TraceLevel:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case NoLevel:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	l.zlog = z
}

func getVariable(v ...any) string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		return fmt.Sprint(v[0])
	}
	return strings.Trim(fmt.Sprint(v...), "]")
}
