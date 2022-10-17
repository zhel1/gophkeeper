package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"gophkeeper/internal/domain"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (r *UserStorage) Create(ctx context.Context, user domain.User) error {
	crUserStmt, err := r.db.PrepareContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id;")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if err := crUserStmt.QueryRowContext(ctx, user.Login, user.Password).Scan(&user.ID); err != nil {
		errCode := err.(*pq.Error).Code
		if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
			return &AlreadyExistsError{Err: domain.ErrUserAlreadyExists}
		}
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *UserStorage) GetByCredentials(ctx context.Context, login, password string) (domain.User, error) {
	user := domain.User{}

	getUserStmt, err := r.db.PrepareContext(ctx, "SELECT id,password FROM users WHERE login=$1;")
	if err != nil {
		return user, &StatementPSQLError{Err: err}
	}
	defer getUserStmt.Close()

	if err := getUserStmt.QueryRowContext(ctx, login).Scan(&user.ID, &user.Password); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user, &NotFoundError{Err: domain.ErrUserNotFound}
		default:
			return user, &ExecutionPSQLError{Err: err}
		}
	}

	if user.Password == password {
		user.Login = login
		return user, nil
	} else {
		return user, domain.ErrUserBadPassword
	}
}

func (r *UserStorage) GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error) {
	user := domain.User{}

	getUserStmt, err := r.db.PrepareContext(ctx, "SELECT id,login,password FROM users WHERE id = (SELECT user_id FROM sessions WHERE refresh_token=$1 and expired_at > now());")
	if err != nil {
		return user, &StatementPSQLError{Err: err}
	}
	defer getUserStmt.Close()

	if err := getUserStmt.QueryRowContext(ctx, refreshToken).Scan(&user.ID, &user.Login, &user.Password); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user, &NotFoundError{Err: domain.ErrUserNotFoundOrSessionWasExpired} //domain.ErrUserBadPassword
		default:
			return user, &ExecutionPSQLError{Err: err}
		}
	}

	return user, nil
}

func (r *UserStorage)SetSession(ctx context.Context, userID int, session domain.Session) error  {
	crUserStmt, err := r.db.PrepareContext(ctx, "INSERT INTO sessions (refresh_token,user_id,expired_at) VALUES ($1, $2, $3);")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if _, err := crUserStmt.ExecContext(ctx, session.RefreshToken, userID, session.ExpiresAt); err != nil {
		errCode := err.(*pq.Error).Code
		if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
			return &AlreadyExistsError{Err: domain.ErrSessionAlreadyExists}
		}
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *UserStorage)UpdateSession(ctx context.Context, userID int, session domain.Session, oldRefreshToken string) error {
	crUserStmt, err := r.db.PrepareContext(ctx, "UPDATE sessions SET refresh_token = $1, user_id = $2, expired_at = $3 WHERE refresh_token = $4;")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if _, err := crUserStmt.ExecContext(ctx, session.RefreshToken, userID, session.ExpiresAt, oldRefreshToken); err != nil {
		errCode := err.(*pq.Error).Code
		if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
			return &AlreadyExistsError{Err: domain.ErrSessionAlreadyExists}
		}
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *UserStorage) Close() error {
	return r.db.Close()
}