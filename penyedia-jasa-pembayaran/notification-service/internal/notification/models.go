package notification

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	EventType    string     `db:"event_type" json:"eventType"`
	EventKey     string     `db:"event_key" json:"eventKey"`
	Payload      string     `db:"payload" json:"payload"` // JSONB stored as string
	CallbackURL  string     `db:"callback_url" json:"callbackUrl"`
	Status       string     `db:"status" json:"status"` // pending, sent, failed
	ResponseCode *int       `db:"response_code" json:"responseCode,omitempty"`
	ResponseBody *string    `db:"response_body" json:"responseBody,omitempty"`
	Attempts     int        `db:"attempts" json:"attempts"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	SentAt       *time.Time `db:"sent_at" json:"sentAt,omitempty"`
}
