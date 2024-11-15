package database

import (
	"context"
	"core/internal/models"
	"database/sql"
	"log"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// GetUserByID retrieves a user by their ID
func (s *service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, display_name FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Name, &user.DisplayName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Load user's credentials
	credentials, err := s.GetCredentialsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = credentials

	return &user, nil
}

// GetUserByName retrieves a user by their username
func (s *service) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, display_name FROM users WHERE name = ?
	`, name).Scan(&user.ID, &user.Name, &user.DisplayName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Load user's credentials
	credentials, err := s.GetCredentialsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = credentials

	return &user, nil
}

// SaveUser saves a new user to the database
func (s *service) SaveUser(ctx context.Context, user *models.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, name, display_name) VALUES (?, ?, ?)
	`, user.ID, user.Name, user.DisplayName)
	return err
}

// SaveCredential saves a new credential to the database
func (s *service) SaveCredential(ctx context.Context, credential *models.Credential) error {
	// Check if the user has any backup-eligible credentials
	existingCreds, err := s.GetCredentialsForUser(ctx, credential.UserID)
	if err != nil {
		return err
	}

	if len(existingCreds) == 0 {
		credential.BackupEligible = true
	} else {
		credential.BackupEligible = false
	}

	recordID := uuid.New().String()

	// Log the credential being saved
	log.Printf("Saving credential: %+v", credential)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO credentials (
			id,
			user_id,
			public_key,
			credential_id,
			sign_count,
			aaguid,
			clone_warning,
			attachment,
			backup_eligible,
			backup_state
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		recordID,
		credential.UserID,
		credential.PublicKey,
		credential.CredentialID,
		credential.SignCount,
		credential.AAGUID,
		credential.CloneWarning,
		string(credential.Attachment),
		credential.BackupEligible,
		credential.BackupState,
	)

	if err != nil {
		log.Printf("Error saving credential: %v", err)
		return err
	}

	log.Printf("Successfully saved credential with ID: %s", recordID)
	return nil
}

// GetCredentialsForUser retrieves all credentials for a given user
func (s *service) GetCredentialsForUser(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			credential_id,
			public_key,
			sign_count,
			aaguid,
			clone_warning,
			attachment,
			backup_eligible,
			backup_state
		FROM credentials
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var cred models.Credential
		var attachmentStr string
		err := rows.Scan(
			&cred.CredentialID,
			&cred.PublicKey,
			&cred.SignCount,
			&cred.AAGUID,
			&cred.CloneWarning,
			&attachmentStr,
			&cred.BackupEligible,
			&cred.BackupState,
		)
		if err != nil {
			return nil, err
		}
		cred.UserID = userID
		cred.Attachment = protocol.AuthenticatorAttachment(attachmentStr)

		// Log the credential details for debugging
		log.Printf("Retrieved credential from DB: %+v", cred)

		wanCred := cred.ToWebauthnCredential()
		log.Printf("Converted to WebAuthn credential: %+v", wanCred)

		credentials = append(credentials, wanCred)
	}
	return credentials, rows.Err()
}

// UpdateCredentialSignCount updates the signCount for a given credential
func (s *service) UpdateCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE credentials
		SET sign_count = ?
		WHERE credential_id = ?
	`, signCount, credentialID)
	return err
}
