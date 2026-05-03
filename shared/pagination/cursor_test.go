package pagination

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncodeDecode(t *testing.T) {
	cursor := Cursor{ID: uuid.New(), CreatedAt: time.Now().UTC()}
	encoded, err := Encode(cursor)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.ID != cursor.ID {
		t.Fatalf("id inatteso: %s", decoded.ID)
	}
}

func TestNormalizeLimit(t *testing.T) {
	if got := NormalizeLimit(0); got != DefaultLimit {
		t.Fatalf("default inatteso: %d", got)
	}
	if got := NormalizeLimit(MaxLimit + 1); got != MaxLimit {
		t.Fatalf("max inatteso: %d", got)
	}
}
