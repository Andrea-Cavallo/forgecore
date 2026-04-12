package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Cursor holds the position for cursor-based pagination.
type Cursor struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// Page holds pagination metadata returned to callers.
type Page struct {
	NextCursor *string `json:"next_cursor,omitempty"`
	PrevCursor *string `json:"prev_cursor,omitempty"`
	HasNext    bool    `json:"has_next"`
	HasPrev    bool    `json:"has_prev"`
	Limit      int     `json:"limit"`
}

// Encode serializes a Cursor to a URL-safe base64 string.
func Encode(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("cursor encode: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Decode deserializes a cursor string back to a Cursor.
func Decode(s string) (Cursor, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor json: %w", err)
	}
	return c, nil
}

// SQL returns the WHERE clause fragment for cursor-based pagination.
// Usage: WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC LIMIT $3
func SQL(cursor *Cursor, limit int) (where string, args []any, queryLimit int) {
	queryLimit = limit + 1 // fetch one extra to detect next page
	if cursor == nil {
		return "", nil, queryLimit
	}
	return "(created_at, id) < ($1, $2)", []any{cursor.CreatedAt, cursor.ID}, queryLimit
}
