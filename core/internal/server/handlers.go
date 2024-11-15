package server

import (
	"core/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

// BeginRegistration starts the registration process
func (s *Server) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" || req.DisplayName == "" {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create a new user
	userID := uuid.New().String()
	user := &models.User{
		ID:          userID,
		Name:        req.Username,
		DisplayName: req.DisplayName,
	}

	// Save the user to the database
	err = s.db.SaveUser(r.Context(), user)
	if err != nil {
		log.Printf("Failed to save user: %v", err)
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	// Begin registration
	options, sessionData, err := s.webAuthn.BeginRegistration(
		user,
		// You can customize options here
	)
	if err != nil {
		log.Printf("Failed to begin registration: %v", err)
		http.Error(w, "Failed to begin registration", http.StatusInternalServerError)
		return
	}

	// Store session data
	s.sessionStore[userID] = sessionData

	// Create response with both options and userID
	response := struct {
		PublicKey *protocol.CredentialCreation `json:"publicKey"`
		UserID    string                       `json:"userID"`
	}{
		PublicKey: options,
		UserID:    userID,
	}

	// Return options to client
	jsonResponse(w, response)
}

// FinishRegistration completes the registration process
func (s *Server) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	// First, get the userID from query parameter
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		userID = r.Header.Get("X-User-ID") // Fallback to header
	}

	if userID == "" {
		log.Printf("UserID not provided")
		http.Error(w, "UserID not provided", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get session data
	sessionData, ok := s.sessionStore[userID]
	if !ok {
		log.Printf("Session data not found for user ID: %s", userID)
		http.Error(w, "Session data not found", http.StatusBadRequest)
		return
	}

	credential, err := s.webAuthn.FinishRegistration(user, *sessionData, r)
	if err != nil {
		log.Printf("Failed to finish registration: %v", err)
		http.Error(w, "Failed to finish registration", http.StatusBadRequest)
		return
	}

	log.Printf("Credential details: %+v", credential)

	// Extract backup flags from the credential's Flags field
	backupEligible := false
	// backupState := true

	// Save the credential with the flags
	cred := &models.Credential{
		UserID:         user.ID,
		PublicKey:      credential.PublicKey,
		CredentialID:   credential.ID,
		SignCount:      credential.Authenticator.SignCount,
		AAGUID:         credential.Authenticator.AAGUID,
		CloneWarning:   credential.Authenticator.CloneWarning,
		Attachment:     credential.Authenticator.Attachment,
		BackupEligible: backupEligible,
		// BackupState:    backupState,
	}

	log.Printf("Saving credential with flags - BackupEligible: %v",
		backupEligible)

	err = s.db.SaveCredential(r.Context(), cred)
	if err != nil {
		log.Printf("Failed to save credential: %v", err)
		http.Error(w, "Failed to save credential", http.StatusInternalServerError)
		return
	}

	// Log successful registration
	log.Printf("Successfully registered credential for user %s", user.ID)

	// Clean up session data
	delete(s.sessionStore, userID)

	jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) BeginLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByName(r.Context(), req.Username)
	if err != nil || user == nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	options, sessionData, err := s.webAuthn.BeginLogin(user)
	if err != nil {
		log.Printf("Failed to begin login: %v", err)
		http.Error(w, "Failed to begin login", http.StatusInternalServerError)
		return
	}

	s.sessionStore[user.ID] = sessionData

	// Create response with both options and userID
	response := struct {
		PublicKey *protocol.CredentialAssertion `json:"publicKey"`
		UserID    string                        `json:"userID"`
	}{
		PublicKey: options,
		UserID:    user.ID,
	}

	jsonResponse(w, response)
}

// FinishLogin completes the login process
func (s *Server) FinishLogin(w http.ResponseWriter, r *http.Request) {
	// Get userID from query parameter
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("UserID not provided")
		http.Error(w, "UserID not provided", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	sessionData, ok := s.sessionStore[user.ID]
	if !ok {
		log.Printf("Session data not found for user ID: %s", user.ID)
		http.Error(w, "Session data not found", http.StatusBadRequest)
		return
	}

	credential, err := s.webAuthn.FinishLogin(user, *sessionData, r)
	if err != nil {
		log.Printf("Login failed with detailed error: %+v", err)
		http.Error(w, "Failed to finish login", http.StatusUnauthorized)
		return
	}

	// Log successful validation
	log.Printf("Successfully validated credential for user %s", user.ID)

	// Update credential's sign count and backup state
	err = s.db.UpdateCredentialSignCount(r.Context(), credential.ID, credential.Authenticator.SignCount)
	if err != nil {
		log.Printf("Failed to update credential: %v", err)
		http.Error(w, "Failed to update credential", http.StatusInternalServerError)
		return
	}

	delete(s.sessionStore, user.ID)

	// Create session for authenticated user
	sessionID := uuid.New().String()
	s.sessionMutex.Lock()
	s.userSessions[sessionID] = user.ID
	s.sessionMutex.Unlock()

	// Set cookie with session ID
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	jsonResponse(w, map[string]string{"status": "ok"})
}

// GetCurrentUser returns the current user's information if authenticated
func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Optionally, you may want to omit sensitive information
	responseUser := struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
	}

	jsonResponse(w, responseUser)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
