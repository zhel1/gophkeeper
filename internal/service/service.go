package service

import (
	"context"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/storage"
	"gophkeeper/pkg/auth"
	"gophkeeper/pkg/hash"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type UserSignUpInput struct {
	Login    string
	Password string
}

type UserSignInInput struct {
	Login    string
	Password string
}

type Users interface {
	SignUp(ctx context.Context, input UserSignUpInput) error
	SignIn(ctx context.Context, input UserSignInInput) (domain.Tokens, error)
	RefreshTokens(ctx context.Context, token string) (domain.Tokens, error)
}

//**********************************************************************************************************************
type Materials interface {
	GetAllTextData(ctx context.Context, userID int) ([]domain.TextData, error)
	UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error
	CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error

	GetAllCardData(ctx context.Context, userID int) ([]domain.CardData, error)
	UpdateCardDataByID(ctx context.Context, userID int, data domain.CardData) error
	CreateNewCardData(ctx context.Context, id int, data domain.CardData) error

	GetAllCredData(ctx context.Context, userID int) ([]domain.CredData, error)
	CreateNewCredData(ctx context.Context, id int, data domain.CredData) error
	UpdateCredDataByID(ctx context.Context, userID int, data domain.CredData) error
}

//**********************************************************************************************************************
type Services struct {
	Users     Users
	Materials Materials
}

type Deps struct {
	Storages        *storage.Storages
	Hasher          hash.PasswordHasher
	TokenManager    auth.TokenManager
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewServices(deps Deps) *Services {
	users := NewUserService(deps.Hasher, deps.Storages.Users, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL)
	materials := NewMaterialsService(deps.Storages.Materials)

	return &Services{
		Users:     users,
		Materials: materials,
	}
}
