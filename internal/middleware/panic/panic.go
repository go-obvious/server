package panic

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog/log"
)

// This is another middleware that must stay on the top since
// we rely on it to convert business-logic-level panics into HTTP 500s.
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr != nil && rvr != http.ErrAbortHandler {
				stack := string(debug.Stack())
				log.Error().
					Str("panic", fmt.Sprint(rvr)).
					Str("host", r.Host).
					Str("method", r.Method).
					Str("uri", r.RequestURI).
					Interface("url", r.URL).
					Str("remote", r.RemoteAddr).
					Strs("stack", strings.Split(stack, "\n")).
					Msg("panicked!")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
