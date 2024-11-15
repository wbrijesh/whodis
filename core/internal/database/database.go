package database

import (
	"context"
	"core/internal/models"
	"database/sql"
	"log"
	"os"

	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

type Service interface {
	Close() error
	CreateTables(ctx context.Context) error

	// User-related methods
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByName(ctx context.Context, name string) (*models.User, error)
	SaveUser(ctx context.Context, user *models.User) error

	// Credential-related methods
	SaveCredential(ctx context.Context, credential *models.Credential) error
	GetCredentialsForUser(ctx context.Context, userID string) ([]webauthn.Credential, error)
	UpdateCredentialSignCount(ctx context.Context, credentialID []byte, signCount uint32) error
}

type service struct {
	db *sql.DB
}

var (
	dburl      = os.Getenv("BLUEPRINT_DB_URL")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}
