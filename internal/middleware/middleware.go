package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %s %s — %v\n%s", r.Method, r.URL.Path, err, debug.Stack())
				hub := sentry.GetHubFromContext(r.Context())
				if hub == nil {
					hub = sentry.CurrentHub().Clone()
				}
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetTag("endpoint", r.URL.Path)
					scope.SetTag("method", r.Method)
					scope.SetLevel(sentry.LevelFatal)
					hub.RecoverWithContext(r.Context(), err)
				})
				hub.Flush(2 * time.Second)
				http.Error(w, "Errore interno del server", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}
