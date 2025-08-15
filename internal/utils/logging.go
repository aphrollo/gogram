package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
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

type Logger struct {
	Level   LogLevel
	Prefix  string
	nocolor bool
	zlog    zerolog.Logger
}

// NewLogger creates a new Logger with prefix
func NewLogger(prefix string) *Logger {
	l := &Logger{
		Prefix: prefix,
	}
	l.initZerolog()
	return l
}

func (l *Logger) NoColor(nocolor ...bool) *Logger {
	if len(nocolor) > 0 {
		l.nocolor = nocolor[0]
	} else {
		l.nocolor = true
	}
	l.initZerolog()
	return l
}

func (l *Logger) Color() bool {
	return !l.nocolor
}

func (l *Logger) SetPrefix(prefix string) *Logger {
	l.Prefix = prefix
	l.initZerolog()
	return l
}

func (l *Logger) Lev() LogLevel {
	return l.Level
}

func (l *Logger) SetLevel(level LogLevel) *Logger {
	l.Level = level
	l.initZerolog()
	return l
}

func (l *Logger) Error(v ...any) {
	if l.Level <= ErrorLevel {
		l.zlog.Error().Str("service", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Warn(v ...any) {
	if l.Level <= WarnLevel {
		l.zlog.Warn().Str("service", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Info(v ...any) {
	if l.Level <= InfoLevel {
		l.zlog.Info().Str("service", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Debug(v ...any) {
	if l.Level <= DebugLevel {
		l.zlog.Debug().Str("service", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Trace(v ...any) {
	if l.Level <= TraceLevel {
		l.zlog.Trace().Str("service", l.Prefix).Msg(getVariable(v...))
	}
}

func (l *Logger) Panic(v ...any) {
	stack := make([]byte, 2048)
	runtime.Stack(stack, false)
	l.zlog.Panic().
		Str("service", l.Prefix).
		Str("stack", string(stack)).
		Msg(getVariable(v...))
}

func (l *Logger) initZerolog() {
	zerolog.TimeFieldFormat = "15:04:05"

	// Cut caller path to module path + version
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		if idx := strings.Index(file, "github.com/"); idx != -1 {
			file = file[idx:]
		}
		return fmt.Sprintf("%s:%d", file, line)
	}

	// Determine level from env or fall back
	envLevel := os.Getenv("LOG_LEVEL")
	if envLevel == "" {
		envLevel = "info"
	}
	level, err := zerolog.ParseLevel(envLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Console writer (color-aware)
	colorableStdout := colorable.NewColorableStdout()
	consoleWriter := zerolog.ConsoleWriter{
		Out:        colorableStdout,
		TimeFormat: zerolog.TimeFieldFormat,
		NoColor:    l.nocolor,
	}

	var writers []io.Writer
	writers = append(writers, consoleWriter)

	// Optional file writer
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			writers = append(writers, file)
		} else {
			fmt.Printf("failed to open log file %s: %v\n", logFile, err)
		}
	}

	multi := io.MultiWriter(writers...)

	// Build logger with timestamp, caller, and prefix
	base := zerolog.New(multi).With().Timestamp().Caller().Str("service", l.Prefix).Logger()

	l.zlog = base
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
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
