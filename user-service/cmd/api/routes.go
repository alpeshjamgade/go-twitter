package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
)

type AuthRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/plug"))

	mux.Post("/user/signup", app.Signup)
	mux.Post("/user/login", app.Login)
	mux.With(app.authenticate).Delete("/user/logout", app.Logout)
	mux.With(app.authenticate).Get("/user/profile", app.UserProfile)

	return mux
}

func (app *Config) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("calling authentication service")

		emailCookie, err := r.Cookie("email")
		if err != nil {
			log.Print("user cookie not present")
			app.errorJSON(w, errors.New("invalid session"))
			return
		}

		token, err := r.Cookie("Authorization")
		if err != nil {
			log.Print("Authorization cookie not present")
			app.errorJSON(w, errors.New("invalid session"))
			return
		}
		_, err = app.validateToken(emailCookie.Value, token.Value)
		if err != nil {
			app.errorJSON(w, errors.New("invalid session"))
			return
		}

		// Validation passed, call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
