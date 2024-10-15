package panic

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
)

// This is another middleware that must stay on the top since
// we rely on it to convert business-logic-level panics into HTTP 500s.
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr != nil && rvr != http.ErrAbortHandler {
				stack := string(debug.Stack())
				logrus.WithFields(logrus.Fields{
					"panic":  fmt.Sprint(rvr),
					"host":   r.Host,
					"method": r.Method,
					"uri":    r.RequestURI,
					"url":    r.URL,
					"remote": r.RemoteAddr,
					"stack":  strings.Split(stack, "\n"),
				}).Error("panicked!")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
