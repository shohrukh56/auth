package authorized

import (
	"context"
	"github.com/shohrukh56/auth/pkg/core/token"
	"log"
	"net/http"
)

func Authorized(roles []string, payload func(ctx context.Context) interface{}) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {

			auth := payload(request.Context()).(*token.Payload)
			for _, role := range roles {
				for _, r := range auth.Roles {
					if role == r {
						log.Printf("access granted %v %v", roles, auth)
						next(writer, request)
						return
					}
				}
			}

			http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}
}
