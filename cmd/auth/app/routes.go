package app

import (
	"github.com/shohrukh56/auth/pkg/core/token"
	"github.com/shohrukh56/auth/pkg/mux/middleware/authenticated"
	"github.com/shohrukh56/auth/pkg/mux/middleware/authorized"
	"github.com/shohrukh56/auth/pkg/mux/middleware/jwt"
	"github.com/shohrukh56/auth/pkg/mux/middleware/logger"
	"reflect"
)

func (s *Server) InitRoutes() {

	s.router.GET(
		"/",
		s.handleIndex(),
		logger.Logger("Index"),
	)


	s.router.POST(
		"/api/tokens",
		s.handleCreateToken(),
		logger.Logger("TOKEN by log/pass"),
	)

	s.router.GET(
		"/api/users",
		s.handleProfile(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USERS/me"),
	)

	s.router.DELETE(
		"/api/users/{id}",
		s.handleDeleteProfile(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		authorized.Authorized([]string{"Admin"}, jwt.FromContext),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USER Delete"),
	)
	s.router.POST(
		"/api/users/{id}",
		s.handleUser(),
		authenticated.Authenticated(jwt.IsContextNonEmpty),
		authorized.Authorized([]string{"Admin"}, jwt.FromContext),
		jwt.JWT(reflect.TypeOf((*token.Payload)(nil)).Elem(), s.secret),
		logger.Logger("USER Register/Update"),
	)
}
