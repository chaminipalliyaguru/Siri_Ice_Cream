package auth

import (
	"strings"
	"sync"
	"time"
)

type TokenBlacklist struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

// NewTokenBlacklist creates an empty blacklist.
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{tokens: make(map[string]time.Time)}
}

// Add puts a token on the blacklist until it would have expired.
func (b *TokenBlacklist) Add(token string, expiresAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.removeExpired()
	// Store a copy of the text. Fiber reuses the memory of a request for the
	// next request, so without the copy the saved token would later change into
	// another request's token (and that new token would look logged out).
	b.tokens[strings.Clone(token)] = expiresAt
}

// Contains reports whether the token has been logged out.
func (b *TokenBlacklist) Contains(token string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, found := b.tokens[token]
	return found
}

// removeExpired deletes tokens that have expired, to keep the map small.
// The caller must hold the lock.
func (b *TokenBlacklist) removeExpired() {
	now := time.Now()
	for token, expiresAt := range b.tokens {
		if now.After(expiresAt) {
			delete(b.tokens, token)
		}
	}
}
