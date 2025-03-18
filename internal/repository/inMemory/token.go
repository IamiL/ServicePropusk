package inMemorySession

import (
	"errors"
	"time"
)

var sessions = map[string]session{}

// each session contains the username of the user and the time at which it expires
type session struct {
	username string
	expiry   time.Time
}

type TokenStore struct {
	sessions map[string]session
}

func New() *TokenStore {
	return &TokenStore{sessions: make(map[string]session)}
}

func (s *TokenStore) Get(token string) (string, error) {
	return "", errors.New("Not implemented")
}
func (s *TokenStore) NewSession(
	uid string,
	sessionToken string,
	expiresAt time.Time,
) error {
	return errors.New("Not implemented")
}
func (s *TokenStore) DeleteSession(uid string) error {
	return errors.New("Not implemented")
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}
