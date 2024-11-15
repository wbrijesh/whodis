package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/joho/godotenv/autoload"

	"core/internal/database"
	"core/internal/models"
)

type Server struct {
	port int
	db   database.Service

	webAuthn     *webauthn.WebAuthn
	sessionStore map[string]*webauthn.SessionData

	userSessions map[string]string
	sessionMutex sync.RWMutex
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dbService := database.New()

	// Initialize WebAuthn with correct config
	wconfig := &webauthn.Config{
		RPDisplayName: "My App",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:3000"},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &[]bool{false}[0],
			UserVerification:   protocol.VerificationPreferred,
		},
		AttestationPreference: protocol.PreferNoAttestation,
		Debug:                 true, // Enable debug logging
	}
	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		log.Fatalf("Failed to create WebAuthn from config: %v", err)
	}

	NewServer := &Server{
		port:         port,
		db:           dbService,
		webAuthn:     webAuthn,
		sessionStore: make(map[string]*webauthn.SessionData),
		userSessions: make(map[string]string),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) getUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		return nil, fmt.Errorf("No session cookie")
	}
	sessionID := cookie.Value

	s.sessionMutex.RLock()
	userID, ok := s.userSessions[sessionID]
	s.sessionMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("Invalid session ID")
	}

	user, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("User not found")
	}
	return user, nil
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromSession(r)
		if err != nil || user == nil {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
