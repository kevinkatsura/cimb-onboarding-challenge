package pagination

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func DecodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
