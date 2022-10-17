package storage

import (
	"context"
	"database/sql"
	"gophkeeper/internal/domain"
)

type Storages struct {
	Users        Users
	Materials	 Materials
}

func NewStorages(db *sql.DB) *Storages {
	return &Storages{
		Users: NewUserStorage(db),
		Materials: NewMaterialsStorage(db),
	}
}

type Users interface {
	Create(ctx context.Context, user domain.User) error
	GetByCredentials(ctx context.Context, login, password string) (domain.User, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error)

	SetSession(ctx context.Context, userID int, session domain.Session) error
	UpdateSession(ctx context.Context, userID int, session domain.Session, oldRefreshToken string) error

	Close() error
}

type Materials interface {
	GetAllTextData(ctx context.Context, userID int) ([]domain.TextData, error)
	UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error
	CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error

	Close() error
}