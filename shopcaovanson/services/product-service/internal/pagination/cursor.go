package pagination

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

func Encode(createdAt time.Time, id uuid.UUID) string {
	payload := fmt.Sprintf("%d|%s", createdAt.UTC().UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func Decode(raw string) (Cursor, error) {
	if raw == "" {
		return Cursor{}, errors.New("empty cursor")
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, errors.New("invalid cursor")
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return Cursor{}, errors.New("invalid cursor")
	}
	nano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Cursor{}, errors.New("invalid cursor")
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return Cursor{}, errors.New("invalid cursor")
	}
	return Cursor{CreatedAt: time.Unix(0, nano).UTC(), ID: id}, nil
}
