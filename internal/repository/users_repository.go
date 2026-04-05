package repository

import (
	"context"
	"errors"

	"github.com/Yashi213/room-booking/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, id uuid.UUID, email, passwordHash string, role domain.UserRole) (domain.User, error)
	GetUserWithPasswordByEmail(ctx context.Context, email string) (domain.User, string, error)
}

type usersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) UsersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) CreateUser(ctx context.Context, id uuid.UUID, email, passwordHash string, role domain.UserRole) (domain.User, error) {
	var u domain.User
	var roleStr string

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, role_id)
		VALUES ($1, $2, $3, (SELECT r.id FROM roles r WHERE r.code = $4))
		RETURNING id, email, (SELECT r.code FROM roles r WHERE r.id = role_id), created_at
	`, id, email, passwordHash, string(role)).Scan(&u.ID, &u.Email, &roleStr, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, ErrConflict
		}
		return domain.User{}, err
	}
	u.Role = domain.UserRole(roleStr)
	return u, nil
}

func (r *usersRepository) GetUserWithPasswordByEmail(ctx context.Context, email string) (domain.User, string, error) {
	var u domain.User
	var hash string
	var roleStr string

	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.password_hash, r.code, u.created_at
		FROM users u
		INNER JOIN roles r ON r.id = u.role_id
		WHERE u.email = $1
	`, email).Scan(&u.ID, &u.Email, &hash, &roleStr, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, "", ErrNotFound
		}
		return domain.User{}, "", err
	}
	u.Role = domain.UserRole(roleStr)
	return u, hash, nil
}
