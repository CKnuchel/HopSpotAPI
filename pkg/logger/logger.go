package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(level string, format string) {
	// Parse level
	var zeroLevel zerolog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		zeroLevel = zerolog.DebugLevel
	case "INFO":
		zeroLevel = zerolog.InfoLevel
	case "WARN":
		zeroLevel = zerolog.WarnLevel
	case "ERROR":
		zeroLevel = zerolog.ErrorLevel
	default:
		zeroLevel = zerolog.InfoLevel
	}

	// Output format
	var output interface{ Write([]byte) (int, error) }
	if strings.ToUpper(format) == "JSON" {
		// Production: Pure JSON (for Loki/Grafana)
		output = os.Stdout
	} else {
		// Development: Human-readable
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	Log = zerolog.New(output).
		Level(zeroLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}

// Shortcut functions
func Debug() *zerolog.Event { return Log.Debug() }
func Info() *zerolog.Event  { return Log.Info() }
func Warn() *zerolog.Event  { return Log.Warn() }
func Error() *zerolog.Event { return Log.Error() }
func Fatal() *zerolog.Event { return Log.Fatal() }
