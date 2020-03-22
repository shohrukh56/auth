package authenticated

import (
	"context"
	"net/http"
)

func Authenticated(isContextEmpty func(ctx context.Context) bool) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			if !isContextEmpty(request.Context()) {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next(writer, request)
		}
	}
}

