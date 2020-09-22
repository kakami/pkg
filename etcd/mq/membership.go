package mq

import (
	"errors"
	"sync"

	"pkg/types"
)

var ErrNoMember = errors.New("no member")

type Membership interface {
	Add(member string)
	Remove(member string)
	Get() (string, error)
	Clear()
}

func NewRoundRobinMembership() Membership {
	return &roundRobin{
		mset: types.NewUnsafeSet(),
	}
}

type roundRobin struct {
	members []string
	mset    types.Set
	idx     int
	mu      sync.RWMutex
}

func (m *roundRobin) Add(member string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mset.Add(member)
	m.members = m.mset.Values()
}

func (m *roundRobin) Remove(member string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mset.Remove(member)
	m.members = m.mset.Values()
}

func (m *roundRobin) Get() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	defer func() {
		m.idx++
	}()
	length := len(m.members)
	if length < 1 {
		return "", ErrNoMember
	}
	return m.members[m.idx%length], nil
}

func (m *roundRobin) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.members = nil
	m.mset = types.NewUnsafeSet()
}
