package inMemory

import "time"

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

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}
