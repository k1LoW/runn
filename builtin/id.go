package builtin

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type ID struct{}

func NewID() *ID {
	return &ID{}
}

// UUIDv4 returns UUID v4.
func (_ *ID) UUIDv4() string {
	return uuid.New().String()
}

// UUIDv6 returns UUID v6.
func (_ *ID) UUIDv6() string {
	return uuid.Must(uuid.NewV6()).String()
}

// UUIDv7 returns UUID v7.
func (_ *ID) UUIDv7() string {
	return uuid.Must(uuid.NewV7()).String()
}

// ULID returns ULID.
func (_ *ID) ULID() string {
	entropy := rand.Reader
	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		panic(err)
	}
	return id.String()
}
