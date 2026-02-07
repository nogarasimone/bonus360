package sentryutil

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func Init() {
	dsn := os.Getenv("SENTRY_DSN")
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      getEnv("SENTRY_ENVIRONMENT", "production"),
		Release:          getEnv("SENTRY_RELEASE", "bonus360@1.0.0"),
		TracesSampleRate: 0.2,
		EnableTracing:    dsn != "",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.User = sentry.User{}
			return event
		},
	})
	if err != nil {
		log.Printf("Sentry init (non-blocking): %s", err)
	}
	if dsn == "" {
		log.Println("SENTRY_DSN vuoto — error tracking disabilitato")
	} else {
		log.Println("Sentry inizializzato")
	}
}

func Flush() { sentry.Flush(2 * time.Second) }

func CaptureError(err error, tags map[string]string) {
	if err == nil {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		sentry.CaptureException(err)
	})
}

func CaptureMessage(msg string, level sentry.Level, tags map[string]string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		sentry.CaptureMessage(msg)
	})
}

// LevelWarning returns sentry.LevelWarning so callers don't need to import sentry-go directly.
func LevelWarning() sentry.Level { return sentry.LevelWarning }

func getEnv(key, fb string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fb
}
