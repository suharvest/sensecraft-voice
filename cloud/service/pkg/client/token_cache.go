package client

import "sync"

// TokenCache
// TODO: 临时实现，后续优化
type TokenCache struct {
	sync.RWMutex
	items map[int64]string
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		items: map[int64]string{},
	}
}

func (s *TokenCache) Get(uid int64) (string, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.items[uid]
	return t, ok
}

func (s *TokenCache) Set(uid int64, token string) {
	s.RLock()
	defer s.RUnlock()

	if s.items == nil {
		s.items = map[int64]string{}
	}
	s.items[uid] = token
}

func (s *TokenCache) Delete(uid int64) {
	s.RLock()
	defer s.RUnlock()

	delete(s.items, uid)
}

func (s *TokenCache) Clear() {
	s.RLock()
	defer s.RUnlock()

	s.items = map[int64]string{}
}
