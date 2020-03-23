package app

import (
	"fmt"
	"github.com/shohrukh56/auth/pkg/core/token"
	"github.com/shohrukh56/auth/pkg/core/user"
	"github.com/shohrukh56/mux/pkg/mux"
	"github.com/shohrukh56/jwt/pkg/jwt"
	"github.com/shohrukh56/rest/pkg/rest"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type (
	TempPath   string
	AssetsPath string
)
type Server struct {
	router        *mux.ExactMux
	pool          *pgxpool.Pool
	secret        jwt.Secret
	tokenSvc      *token.Service
	userSvc       *user.Service
	templatesPath TempPath
	assetsPath    AssetsPath
}

func NewServer(router *mux.ExactMux, pool *pgxpool.Pool, secret jwt.Secret, tokenSvc *token.Service, userSvc *user.Service, templatesPath TempPath, assetsPath AssetsPath) *Server {
	return &Server{router: router, pool: pool, secret: secret, tokenSvc: tokenSvc, userSvc: userSvc, templatesPath: templatesPath, assetsPath: assetsPath}
}

func (s *Server) Start() {
	s.InitRoutes()
}

func (s *Server) Stop() {
}

type ErrorDTO struct {
	Errors []string `json:"errors"`
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleCreateToken() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body token.RequestDTO
		err := rest.ReadJSONBody(request, &body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.json_invalid"},
			})
			log.Print(err)
			return
		}

		response, err := s.tokenSvc.Generate(request.Context(), &body, s.pool)

		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err2 := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.password_mismatch", err.Error()},
			})
			if err2 != nil {
				log.Print(err2)
			}
			return
		}

		err = rest.WriteJSONBody(writer, &response)
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) handleDeleteProfile() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		context, ok := mux.FromContext(request.Context(), "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(context)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		profile, err := s.userSvc.Profile(request.Context())
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if int64(id) == profile.Id {
			writer.WriteHeader(http.StatusBadRequest)
			err = rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"you can't delete yourself"},
			})
			if err != nil {
				log.Print(err)
			}
			return
		}
		err = s.userSvc.DelUserByID(request.Context(), int64(id))
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err2 := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{err.Error()},
			})
			if err2 != nil {
				log.Print(err2)
			}
			return
		}

	}
}

func (s *Server) handleIndex() http.HandlerFunc {

	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(
		filepath.Join("web/templates", "index.gohtml"),
	)
	if err != nil {
		panic(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		// executes in many goroutines
		// TODO: fetch data from multiple upstream services
		err = tpl.Execute(writer, struct{ Title string }{Title: "auth",})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}

}
//
//func (s *Server) handleRegister() http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		get := request.Header.Get("Content-Type")
//		fmt.Println(get)
//		if get != "application/json" {
//			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
//			return
//		}
//
//		var newUser token.RequestDTO
//
//		err := rest.ReadJSONBody(request, &newUser)
//		if err != nil {
//			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
//			return
//		}
//		err = s.userSvc.RegisterUser(request.Context(), newUser, s.pool)
//		if err != nil {
//			writer.Write([]byte(err.Error()))
//			return
//		}
//		writer.Write([]byte("done!"))
//
//	}
//}

func (s *Server) handleProfile() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		response, err := s.userSvc.Profile(request.Context())
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			err2 := rest.WriteJSONBody(writer, &ErrorDTO{
				[]string{"err.bad_request"},
			})
			log.Print(err2)
			return
		}
		err = rest.WriteJSONBody(writer, &response)
		if err != nil {
			log.Print(err)
		}

	}
}

func (s *Server) handleUser() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		context, ok := mux.FromContext(request.Context(), "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(context)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		get := request.Header.Get("Content-Type")
		fmt.Println(get)
		if get != "application/json" {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var newUser token.RequestDTO

		err = rest.ReadJSONBody(request, &newUser)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if id == 0 {

			err = s.userSvc.RegisterUser(request.Context(), newUser)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				err2 := rest.WriteJSONBody(writer, &ErrorDTO{
					[]string{err.Error()},
				})
				log.Print(err2)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if id > 0 {
			err := s.userSvc.Update(request.Context(), int64(id), newUser)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				err2 := rest.WriteJSONBody(writer, &ErrorDTO{
					[]string{err.Error()},
				})
				log.Print(err2)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(writer,http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

