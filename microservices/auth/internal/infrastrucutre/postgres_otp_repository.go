package infrastructure

import (
	"context"
	"database/sql"
	"midaslabs/microservices/auth/internal/domain"
	"time"

	"github.com/jmoiron/sqlx"
)

// PostgresOTPRepository implements the OTPRepository interface using PostgreSQL.
type PostgresOTPRepository struct {
	db *sqlx.DB
}

// NewPostgresOTPRepository creates a new PostgresOTPRepository.
func NewPostgresOTPRepository(db *sqlx.DB) *PostgresOTPRepository {
	return &PostgresOTPRepository{db: db}
}

func (r *PostgresOTPRepository) StoreOTP(ctx context.Context, phoneNumber, otp string, expiration time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO otps (phone_number, code, expiration) VALUES ($1, $2, $3)`,
		phoneNumber, otp, expiration)
	return err
}

func (r *PostgresOTPRepository) GetOTP(ctx context.Context, phoneNumber string) (otp string, expiration time.Time, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT code, expiration FROM otps WHERE phone_number = $1`, phoneNumber)
	err = row.Scan(&otp, &expiration)
	if err == sql.ErrNoRows {
		return "", time.Time{}, domain.ErrOTPNotFound
	}
	return otp, expiration, err
}

func (r *PostgresOTPRepository) DeleteOTP(ctx context.Context, phoneNumber string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM otps WHERE phone_number = $1`, phoneNumber)
	return err
}
