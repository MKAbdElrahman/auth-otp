package infrastructure

import (
	"context"
	"midaslabs/microservices/auth/internal/domain"

	"github.com/jmoiron/sqlx"
)

// PostgresActivityRepository implements the ActivityRepository interface using PostgreSQL.
type PostgresActivityRepository struct {
	db *sqlx.DB
}

// NewPostgresActivityRepository creates a new PostgresActivityRepository.
func NewPostgresActivityRepository(db *sqlx.DB) *PostgresActivityRepository {
	return &PostgresActivityRepository{db: db}
}

func (r *PostgresActivityRepository) RecordActivity(ctx context.Context, activity *domain.Activity) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO activities (phone_number, type, timestamp) VALUES ($1, $2, $3)`,
		activity.PhoneNumber, activity.Type, activity.Timestamp)
	return err
}
