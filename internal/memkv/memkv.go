package memkv

import (
	"sync"
	"time"
)

type item struct {
	value     interface{}
	expiresAt time.Time
	version   uint64
}

type InMemoryStore struct {
	data              map[string]item
	defaultExpiration time.Duration
	mu                sync.RWMutex
	stopJanitor       chan struct{}
	nextVersion       uint64
}

var Store *InMemoryStore

func init() {
	Store = NewStore(10 * time.Minute)
}

func NewStore(defaultExpiration time.Duration) *InMemoryStore {
	s := &InMemoryStore{
		data:              make(map[string]item),
		defaultExpiration: defaultExpiration,
		stopJanitor:       make(chan struct{}),
	}

	go s.janitor(1 * time.Minute)

	return s
}

// Set assigns a new store-wide version even when the value is unchanged.
func (s *InMemoryStore) Set(key string, value interface{}, expiration ...time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp := s.defaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	var expiresAt time.Time
	if exp > 0 {
		expiresAt = time.Now().Add(exp)
	}

	s.nextVersion++
	if s.nextVersion == 0 {
		panic("memkv version exhausted")
	}
	s.data[key] = item{
		value:     value,
		expiresAt: expiresAt,
		version:   s.nextVersion,
	}
}

// Get reads a live value and expires old entries atomically.
func (s *InMemoryStore) Get(key string) (interface{}, bool) {
	value, _, ok := s.GetVersioned(key)
	return value, ok
}

// GetVersioned identifies the exact observed entry for conditional deletion.
func (s *InMemoryStore) GetVersioned(key string) (interface{}, uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, found := s.data[key]

	if !found {
		return nil, 0, false
	}

	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		delete(s.data, key)
		return nil, 0, false
	}

	return it.value, it.version, true
}

// DeleteVersion removes only the observed entry, never a later replacement.
func (s *InMemoryStore) DeleteVersion(key string, version uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if it, ok := s.data[key]; !ok || version == 0 || it.version != version {
		return false
	}
	delete(s.data, key)
	return true
}

func (s *InMemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *InMemoryStore) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopJanitor:
			return
		}
	}
}

func (s *InMemoryStore) cleanup() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, it := range s.data {
		if !it.expiresAt.IsZero() && now.After(it.expiresAt) {
			delete(s.data, k)
		}
	}
}

func (s *InMemoryStore) Close() {
	close(s.stopJanitor)
}
