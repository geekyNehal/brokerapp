package user

import (
	"context"
	"database/sql"
	"log"
	"time"

	"brokerapp/internal/db"
)

type MySQLRepository struct {
	db *db.MySQL
}

func NewMySQLRepository(db *db.MySQL) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password, created_at)
		VALUES (?, ?, ?)
	`

	_, err := r.db.Exec(ctx, query,
		user.Email,
		user.Password,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *MySQLRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password, created_at
		FROM users
		WHERE email = ?
	`

	row := r.db.QueryRow(ctx, query, email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *MySQLRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, email, password, created_at
		FROM users
		WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *MySQLRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = ?
	`

	log.Printf("Deleting all refresh tokens for user %d", userID)
	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("Error deleting all refresh tokens: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Deleted %d refresh token(s) for user %d", rowsAffected, userID)
	return nil
}

func (r *MySQLRepository) StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	// First delete all existing refresh tokens for this user
	if err := r.DeleteAllRefreshTokens(ctx, userID); err != nil {
		return err
	}

	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`

	log.Printf("Storing refresh token for user %d: %s", userID, token)
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
		return err
	}

	return nil
}

func (r *MySQLRepository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT user_id, token, expires_at
		FROM refresh_tokens
		WHERE token = ?
	`

	log.Printf("Getting refresh token: %s", token)
	row := r.db.QueryRow(ctx, query, token)
	refreshToken := &RefreshToken{}
	err := row.Scan(&refreshToken.UserID, &refreshToken.Token, &refreshToken.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Refresh token not found: %s", token)
			return nil, ErrRefreshTokenNotFound
		}
		log.Printf("Error getting refresh token: %v", err)
		return nil, err
	}

	log.Printf("Found refresh token for user %d, expires at: %v", refreshToken.UserID, refreshToken.ExpiresAt)
	return refreshToken, nil
}

func (r *MySQLRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = ?
	`

	log.Printf("Deleting refresh token: %s", token)
	result, err := r.db.Exec(ctx, query, token)
	if err != nil {
		log.Printf("Error deleting refresh token: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Deleted %d refresh token(s)", rowsAffected)
	return nil
}
