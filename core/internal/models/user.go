package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents a user in our system.
type User struct {
    ID             string                 // Unique identifier for the user
    Name           string                 // Username
    DisplayName    string                 // Full name or display name
    Credentials    []webauthn.Credential  // WebAuthn credentials
}

// Ensure User satisfies the webauthn.User interface
var _ webauthn.User = &User{}

// WebAuthnID returns the user's unique ID
func (u *User) WebAuthnID() []byte {
    return []byte(u.ID)
}

// WebAuthnName returns the user's username
func (u *User) WebAuthnName() string {
    return u.Name
}

// WebAuthnDisplayName returns the user's display name
func (u *User) WebAuthnDisplayName() string {
    return u.DisplayName
}

// WebAuthnIcon returns the user's icon URL (optional)
func (u *User) WebAuthnIcon() string {
    return ""
}

// WebAuthnCredentials returns the user's credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
    return u.Credentials
}
