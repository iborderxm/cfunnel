package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

var (
	ErrInvalidUUIDFormat = errors.New("invalid UUID format")
)

type UUID [16]byte

func NewRandom() (UUID, error) {
	var u UUID
	if _, err := rand.Read(u[:]); err != nil {
		return UUID{}, err
	}
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80
	return u, nil
}

func Parse(s string) (UUID, error) {
	var u UUID
	if len(s) != 36 {
		return UUID{}, ErrInvalidUUIDFormat
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return UUID{}, ErrInvalidUUIDFormat
	}
	hexStr := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	if _, err := hex.Decode(u[:], []byte(hexStr)); err != nil {
		return UUID{}, ErrInvalidUUIDFormat
	}
	return u, nil
}

func FromBytes(b []byte) (UUID, error) {
	if len(b) != 16 {
		return UUID{}, ErrInvalidUUIDFormat
	}
	var u UUID
	copy(u[:], b)
	return u, nil
}

func (u UUID) String() string {
	return hex.EncodeToString(u[0:4]) + "-" +
		hex.EncodeToString(u[4:6]) + "-" +
		hex.EncodeToString(u[6:8]) + "-" +
		hex.EncodeToString(u[8:10]) + "-" +
		hex.EncodeToString(u[10:16])
}

func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *UUID) UnmarshalText(b []byte) error {
	parsed, err := Parse(string(b))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}
