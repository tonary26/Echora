package database

import (
	"context"
	"echora/internal/models"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore struct {
	db *pgxpool.Pool
}

func NewAuthStore(db *pgxpool.Pool) *AuthStore {
	return &AuthStore{db: db}
}

func (s *AuthStore) Register(ctx context.Context, input models.UserCreateInput) (*models.User, error) {
	var user models.User

	query := `
		INSERT INTO users 
		(name, email, avatar_url, password_hash) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, avatar_url, created_at;
	`

	err := s.db.QueryRow(ctx, query, input.Name, input.Email, input.Avatar_url, input.Password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Avatar_url,
		&user.Created_at,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, name, email, COALESCE(avatar_url, ''), password_hash, created_at, updated_at
		FROM users 
		WHERE email = $1;
	`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Avatar_url,
		&user.Password_hash,
		&user.Created_at,
		&user.Updated_at,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (s *AuthStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, name, email, COALESCE(avatar_url, ''), created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Avatar_url,
		&user.Created_at,
		&user.Updated_at,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
