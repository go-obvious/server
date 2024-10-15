package apicaller

import (
	"context"
	"errors"
	"net/http"
)

var ErrMissingContext = errors.New("missing context")

type ctxKeyType int

const (
	CtxKey        ctxKeyType = 1
	APIVersionHdr string     = "APIVersion"
)

type Context struct {
	UserAgent  string `json:"user_agent"`
	APIVersion string `json:"api_version"`
}

func NewContext(r *http.Request) *Context {
	ref := Context{
		UserAgent: r.UserAgent(),
	}

	ref.APIVersion = r.Header.Get(APIVersionHdr)

	return &ref
}

func GetContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}

	if thisCtx, ok := ctx.Value(CtxKey).(*Context); ok {
		return thisCtx
	}

	return nil
}

func SaveContext(ctx context.Context, ref *Context) context.Context {
	return context.WithValue(ctx, CtxKey, ref)
}

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := SaveContext(r.Context(), NewContext(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
