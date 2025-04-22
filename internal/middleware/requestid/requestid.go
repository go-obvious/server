package requestid

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type ctxKeyType int

const (
	CtxKey ctxKeyType = iota
)

type Context struct {
	RequestID string `json:"request_id"`
}

func NewContext(r *http.Request) *Context {
	return &Context{
		RequestID: middleware.GetReqID(r.Context()),
	}
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
		reqID := middleware.GetReqID(r.Context())
		if reqID == "" {
			reqID = middleware.RequestIDHeader
		}
		ctx := SaveContext(r.Context(), &Context{RequestID: reqID})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return middleware.RequestID(http.HandlerFunc(fn))
}
