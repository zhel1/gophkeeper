package service

import (
	"context"
	"errors"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/storage"
	"gophkeeper/pkg/auth"
	"gophkeeper/pkg/hash"
	"strconv"
	"time"
)

type UserService struct {
	hasher      	hash.PasswordHasher
	storage 		storage.Users
	tokenManager 	auth.TokenManager
	accessTokenTTL	time.Duration
	refreshTokenTTL	time.Duration
}

func NewUserService(h hash.PasswordHasher, s storage.Users, tm auth.TokenManager, at time.Duration, rt time.Duration) *UserService {
	return &UserService{
		hasher: h,
		storage: s,
		tokenManager: tm,
		accessTokenTTL:	at,
		refreshTokenTTL: rt,
	}
}

func (s *UserService) SignUp(ctx context.Context, input UserSignUpInput) error {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return err
	}

	user := domain.User{
		Login: input.Login,
		Password: passwordHash,
	}

	if err := s.storage.Create(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *UserService) SignIn(ctx context.Context, input UserSignInInput) (domain.Tokens, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return domain.Tokens{}, err
	}

	user, err := s.storage.GetByCredentials(ctx, input.Login, passwordHash)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.Tokens{}, err
		}

		return domain.Tokens{}, err
	}

	return s.createSession(ctx, user.ID, "")
}

//TODO add old tokens to the black list until thea are spoiled
func (s *UserService) RefreshTokens(ctx context.Context, token string) (domain.Tokens, error) {
	user, err := s.storage.GetByRefreshToken(ctx, token)
	if err != nil {
		return domain.Tokens{}, err
	}

	return s.createSession(ctx, user.ID, token)
}

func (s *UserService) createSession(ctx context.Context, userID int, token string) (domain.Tokens, error) {
	var (
		res domain.Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.NewJWT(strconv.Itoa(userID), s.accessTokenTTL)
	if err != nil {
		return res, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return res, err
	}

	session := domain.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}

	if token == "" {
		err = s.storage.SetSession(ctx, userID, session)
	} else {
		err = s.storage.UpdateSession(ctx, userID, session, token)
	}

	return res, err
}
