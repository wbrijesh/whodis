package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Your frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Important for cookies
		MaxAge:           300,  // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", s.HelloWorldHandler)

	// Registration endpoints
	r.Post("/register/begin", s.BeginRegistration)
	r.Post("/register/finish", s.FinishRegistration)

	// Login endpoints
	r.Post("/login/begin", s.BeginLogin)
	r.Post("/login/finish", s.FinishLogin)

	// Protected endpoint
	r.With(s.AuthMiddleware).Get("/me", s.GetCurrentUser)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
