package web

import (
	"errors"
	"hestia/pkg/auth"
	"hestia/pkg/custerrors"
	"hestia/pkg/middlewares"
	"hestia/pkg/services"
	"log/slog"
	"net/http"
)

// ServerDeps are the dependencies for the server.
type ServerDeps struct {
	Logger       *slog.Logger
	UserService  *services.UserService
	FlatService  *services.FlatService
	AuthService  *services.AuthService
	EmailService *services.EmailService
	JWT          *auth.JWTConfig
	Interceptor  *auth.Interceptor
}

func NewServer(s *ServerDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		err := s.AuthService.Login(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		err := s.AuthService.Register(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("GET /api/v1/users", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.UserService.GetAll(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("GET /api/v1/users/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.UserService.Get(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("POST /api/v1/users", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.UserService.Post(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	mux.Handle("PUT /api/v1/users/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.UserService.Put(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.Handle("DELETE /api/v1/users/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.UserService.Delete(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	mux.Handle("GET /api/v1/flats", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.FlatService.GetAll(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("GET /api/v1/flats/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.FlatService.Get(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("POST /api/v1/flats", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.FlatService.Post(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	mux.Handle("PUT /api/v1/flats/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.FlatService.Put(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.Handle("DELETE /api/v1/flats/{id}", middlewares.CheckJWT(s.Interceptor, func(w http.ResponseWriter, r *http.Request) {
		err := s.FlatService.Delete(w, r)
		if err != nil {
			s.handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	return mux
}

func (s *ServerDeps) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, custerrors.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}
