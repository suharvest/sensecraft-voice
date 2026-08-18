package client

import "sync"

type UserCache struct {
	sync.RWMutex
	items map[int64]int
}

func NewUserCache() *UserCache {
	return &UserCache{
		items: map[int64]int{},
	}
}

func (s *UserCache) Get(uid int64) (int, bool) {
	s.RLock()
	defer s.RUnlock()

	status, ok := s.items[uid]
	return status, ok
}

func (s *UserCache) Set(uid int64, status int) {
	s.RLock()
	defer s.RUnlock()

	if s.items == nil {
		s.items = map[int64]int{}
	}
	s.items[uid] = status
}

func (s *UserCache) Delete(uid int64) {
	s.RLock()
	defer s.RUnlock()

	delete(s.items, uid)
}

func (s *UserCache) Clear() {
	s.RLock()
	defer s.RUnlock()

	s.items = map[int64]int{}
}
