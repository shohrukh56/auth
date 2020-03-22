package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/shohrukh56/auth/pkg/core/token"
	"github.com/shohrukh56/auth/pkg/mux/middleware/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Start() {

	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		panic(errors.New("can't create database"))
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), `
CREATE TABLE if not exists users (
   id BIGSERIAL PRIMARY KEY,
   username TEXT NOT NULL unique,
   password TEXT NOT NULL,
   admin BOOLEAN DEFAULT FALSE,
   removed BOOLEAN DEFAULT FALSE
);
`)
	if err != nil {
		panic(errors.New("can't create database"))
	}
	_, err = conn.Exec(context.Background(), `
Insert into users(username, password, admin) Values ('shohrukh', '$2a$10$yh.tFQKJH6xYTU4ZijsdZe0fzRZvzQzVP6Opd616dxvSdEwQ18tt2', True) on conflict do nothing;
`)
	if err != nil {
		panic(errors.New("can't create database"))
	}

}

type ResponseDTO struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (s *Service) Profile(ctx context.Context) (response ResponseDTO, err error) {
	auth, ok := jwt.FromContext(ctx).(*token.Payload)
	if !ok {
		return ResponseDTO{}, errors.New("bad request")
	}

	return ResponseDTO{
		Id:     auth.Id,
		Name:   auth.Username,
		Avatar: "https://i.pravatar.cc/50",
	}, nil
}

func (s *Service) FindUserByID(ctx context.Context, id int64) (response ResponseDTO, err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	var (
		username string
		isAdmin  bool
		removed  bool
	)
	fmt.Println(id)
	err = conn.QueryRow(ctx, `select username, admin, removed from users where id = $1;`, id).Scan(&username, &isAdmin, &removed)
	if err != nil {
		err = errors.New("no such user")
		return
	}
	if isAdmin {
		err = errors.New("you can't delete admin")
		return
	}
	if removed {
		err = errors.New("this user already deleted")
		return
	}

	return ResponseDTO{
		Id:     id,
		Name:   username,
		Avatar: "https://i.pravatar.cc/50",
	}, nil
}

func (s *Service) DelUserByID(ctx context.Context, id int64) (err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, `UPDATE users SET removed = True WHERE id = $1 and admin = false;`, id)
	if err != nil {
		err = errors.New("server error")
		return
	}
	return
}

func (s *Service) RegisterUser(ctx context.Context, newUser token.RequestDTO) (err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()
	password, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		err = errors.New("server error")
		return
	}
	_, err = conn.Exec(ctx, `insert into users(username, password) Values ($1, $2);`, newUser.Username, password)
	if err != nil {
		err = errors.New("server error")
		return
	}
	return
}

func (s *Service) Update(ctx context.Context, id int64, dto token.RequestDTO) (err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer conn.Release()

	begin, err := conn.Begin(ctx)
	if err != nil {
		err = errors.New("server error")
		return
	}
	defer func() {
		if err != nil {
			err2 := begin.Rollback(ctx)
			if err2 != nil {
				log.Printf("can't rollback %v", err2)
			}
			return
		}
		err2 := begin.Commit(ctx)
		if err2 != nil {
			log.Printf("can't commit %v", err2)
		}

	}()
	if dto.Username != "" {
		_, err = begin.Exec(ctx, `UPDATE users SET username = $2 WHERE id = $1;`, id, dto.Username)
		if err != nil {
			return
		}
	}
	if dto.Password != "" {
		password, err2 := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
		if err2 != nil {
			err = errors.New("server error")
			return
		}
		_, err = begin.Exec(ctx, `UPDATE users SET password = $2 WHERE id = $1;`, id, password)
		if err != nil {
			return
		}
	}
		_, err = begin.Exec(ctx, `UPDATE users SET admin = $2 WHERE id = $1;`, id, dto.Admin)
		if err != nil {
			return
		}
	return
}
