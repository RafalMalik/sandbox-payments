package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rmalik/sandbox-payments/internal/payment"
)

// SQLiteStore implements payment.Store backed by SQLite.
type SQLiteStore struct {
	db *sql.DB
}

var _ payment.Store = (*SQLiteStore)(nil)

// OpenSQLite opens (or creates) a SQLite database and runs migrations.
func OpenSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS payments (
	id TEXT PRIMARY KEY,
	amount INTEGER NOT NULL,
	currency TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	payment_method TEXT NOT NULL DEFAULT '',
	success_url TEXT NOT NULL,
	cancel_url TEXT NOT NULL,
	failed_url TEXT NOT NULL,
	webhook_url TEXT NOT NULL DEFAULT '',
	itn_url TEXT NOT NULL DEFAULT '',
	itn_delay_seconds INTEGER NOT NULL DEFAULT 0,
	metadata TEXT NOT NULL DEFAULT '{}',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	for _, stmt := range []string{
		`ALTER TABLE payments ADD COLUMN itn_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE payments ADD COLUMN itn_delay_seconds INTEGER NOT NULL DEFAULT 0`,
	} {
		_, _ = s.db.Exec(stmt)
	}
	return nil
}

// Close releases the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Create inserts a new payment record.
func (s *SQLiteStore) Create(ctx context.Context, p *payment.Payment) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO payments (
	id, amount, currency, description, status, payment_method,
	success_url, cancel_url, failed_url, webhook_url, itn_url, itn_delay_seconds, metadata,
	created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Amount, p.Currency, p.Description, string(p.Status), p.PaymentMethod,
		p.SuccessURL, p.CancelURL, p.FailedURL, p.WebhookURL, p.ITNURL, p.ITNDelay, p.Metadata,
		p.CreatedAt.UTC().Format(time.RFC3339), p.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// GetByID returns a payment by its identifier.
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (*payment.Payment, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, amount, currency, description, status, payment_method,
	success_url, cancel_url, failed_url, webhook_url, itn_url, itn_delay_seconds, metadata,
	created_at, updated_at
FROM payments WHERE id = ?`, id)

	p, err := scanPayment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	return p, err
}

// Update persists changes to an existing payment.
func (s *SQLiteStore) Update(ctx context.Context, p *payment.Payment) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE payments SET
	amount = ?, currency = ?, description = ?, status = ?, payment_method = ?,
	success_url = ?, cancel_url = ?, failed_url = ?, webhook_url = ?,
	itn_url = ?, itn_delay_seconds = ?, metadata = ?,
	updated_at = ?
WHERE id = ?`,
		p.Amount, p.Currency, p.Description, string(p.Status), p.PaymentMethod,
		p.SuccessURL, p.CancelURL, p.FailedURL, p.WebhookURL, p.ITNURL, p.ITNDelay, p.Metadata,
		p.UpdatedAt.UTC().Format(time.RFC3339), p.ID,
	)
	return err
}

func scanPayment(row *sql.Row) (*payment.Payment, error) {
	var p payment.Payment
	var status, createdAt, updatedAt string

	err := row.Scan(
		&p.ID, &p.Amount, &p.Currency, &p.Description, &status, &p.PaymentMethod,
		&p.SuccessURL, &p.CancelURL, &p.FailedURL, &p.WebhookURL, &p.ITNURL, &p.ITNDelay, &p.Metadata,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	p.Status = payment.Status(status)
	p.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	p.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
