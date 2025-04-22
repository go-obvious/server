package version

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/go-obvious/server/request"
)

type ServerVersion struct {
	Revision string `json:"revision"`
	Tag      string `json:"tag"`
	Time     string `json:"time"`
}

var (
	once                = sync.Once{}
	info *ServerVersion = &ServerVersion{
		Revision: "latest",
		Tag:      "latest",
		Time:     "latest",
	}
)

func SetVersion(i *ServerVersion) {
	once.Do(func() {
		info = i
	})
}

func Endpoint() http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		request.Reply(r, w, info, http.StatusOK)
	})
	return r
}
