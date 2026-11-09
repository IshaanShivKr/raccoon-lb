package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader && (code >= http.StatusOK || code == http.StatusSwitchingProtocols) {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	n, err := r.ResponseWriter.Write(b)
	r.bytes += int64(n)
	return n, err
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			level := slog.LevelInfo
			if rec.status >= http.StatusInternalServerError {
				level = slog.LevelWarn
			}

			logger.Log(r.Context(), level, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"response_bytes", rec.bytes,
				"duration_ms", float64(time.Since(start))/float64(time.Millisecond),
			)
		}()

		next.ServeHTTP(rec, r)
	})
}
