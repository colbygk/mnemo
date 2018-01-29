package core

import (
	"github.com/codegangsta/negroni"
	"io"
	"log"
	"net/http"
	"time"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// Logger inherits from log.Logger used to log messages with the Logger middleware
	*log.Logger
}

// NewLogger returns a new Logger instance
func NewLogger(w io.Writer) *Logger {
	return &Logger{log.New(w, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	end := time.Since(start)

	l.Printf("{\"client\":\"%v\", \"method\":\"%s\" \"url\":\"%s\" \"statusnum\":\"%v\" \"statustext\":\"%s\" \"time\":\"%v\", \"tls\":\"%v\"}", r.RemoteAddr, r.Method, r.URL.Path, res.Status(), http.StatusText(res.Status()), end, r.TLS.HandshakeComplete)
}
