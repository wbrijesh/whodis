package models

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Credential represents a WebAuthn credential
type Credential struct {
	ID             string // database record ID
	UserID         string // foreign key to users table
	PublicKey      []byte // stored public key
	CredentialID   []byte // WebAuthn credential ID
	SignCount      uint32
	AAGUID         []byte
	CloneWarning   bool
	Attachment     protocol.AuthenticatorAttachment
	BackupEligible bool
	BackupState    bool
}

type CredentialFlags struct {
	// Flag UP indicates the users presence.
	UserPresent bool `json:"userPresent"`

	// Flag UV indicates the user performed verification.
	UserVerified bool `json:"userVerified"`

	// Flag BE indicates the credential is able to be backed up and/or sync'd between devices. This should NEVER change.
	BackupEligible bool `json:"backupEligible"`

	// Flag BS indicates the credential has been backed up and/or sync'd. This value can change but it's recommended
	// that RP's keep track of this value.
	BackupState bool `json:"backupState"`
}

// ToWebauthnCredential converts our Credential to a webauthn.Credential
func (c *Credential) ToWebauthnCredential() webauthn.Credential {
	return webauthn.Credential{
		ID:        c.CredentialID,
		PublicKey: c.PublicKey,
		Flags: webauthn.CredentialFlags{
			UserPresent:    true,
			UserVerified:   true,
			BackupEligible: c.BackupEligible,
			BackupState:    c.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			SignCount:    c.SignCount,
			AAGUID:       c.AAGUID,
			CloneWarning: c.CloneWarning,
			Attachment:   c.Attachment,
		},
	}
}
