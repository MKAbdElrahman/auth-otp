package infrastructure

import (
	"context"
	"database/sql"
	"midaslabs/microservices/auth/internal/domain"

	"github.com/jmoiron/sqlx"
)

// PostgresUserRepository implements the UserRepository interface using PostgreSQL.
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetUser(ctx context.Context, phoneNumber string) (*domain.User, error) {
	var user domain.User
	row := r.db.QueryRowContext(ctx, `SELECT phone_number, verified, created_at, updated_at FROM users WHERE phone_number = $1`, phoneNumber)
	if err := row.Scan(&user.PhoneNumber, &user.Verified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) AddUser(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (phone_number, verified, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		user.PhoneNumber, user.Verified, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET verified = $1, updated_at = $2 WHERE phone_number = $3`,
		user.Verified, user.UpdatedAt, user.PhoneNumber)
	return err
}
