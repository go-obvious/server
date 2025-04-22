package healthz

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-obvious/server/healthz"
	"github.com/go-obvious/server/request"
)

func Endpoint() http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := healthz.NewHealthz().Run(); err != nil {
			request.Reply(r, w,
				request.Result{
					Success: false,
					Error:   err.Error(),
				},
				http.StatusServiceUnavailable,
			)
			return
		}
		request.Reply(r, w, request.NewResult(), http.StatusOK)
	})
	return r
}
