package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, n *Notification) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO notification.notifications
		(id, event_type, event_key, payload, callback_url, status, attempts)
		VALUES (:id, :event_type, :event_key, :payload, :callback_url, :status, :attempts)`, n)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, responseCode int, responseBody string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notification.notifications
		 SET status = $1, response_code = $2, response_body = $3, attempts = attempts + 1,
		     sent_at = CASE WHEN $1::varchar = 'sent' THEN NOW() ELSE sent_at END,
		     id = id
		 WHERE id = $4`, status, responseCode, responseBody, id)
	return err
}
