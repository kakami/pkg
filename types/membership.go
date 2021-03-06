package types

import (
	"sync"

	"github.com/pkg/errors"
)

var ErrNoMember = errors.New("no member")

type Membership interface {
	Add(member string)
	Adds(members ...string)
	Members() []string
	Remove(member string)
	Get(key *string) (string, error)
	Clear()
}

func NewRoundRobinMembership() Membership {
	return &roundRobin{
		mset: NewUnsafeSet(),
	}
}

type roundRobin struct {
	members []string
	mset    Set
	idx     int
	mu      sync.Mutex
}

func (m *roundRobin) Add(member string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mset.Add(member)
	m.members = m.mset.Values()
}

func (m *roundRobin) Adds(members ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for idx, _ := range members {
		m.mset.Add(members[idx])
	}
	m.members = m.mset.Values()
}

func (m *roundRobin) Members() []string {
	p := m.members
	return p
}

func (m *roundRobin) Remove(member string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mset.Remove(member)
	m.members = m.mset.Values()
}

func (m *roundRobin) Get(_ *string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	m.mset = NewUnsafeSet()
	m.members = m.mset.Values()
}

// NewConsistentMembership
// ================================================
func NewConsistentMembership(num int) Membership {
	m := &consistent{
		c: NewConsistent(),
	}
	m.c.NumberOfReplicas = num
	return m
}

type consistent struct {
	c *Consistent
}

func (m *consistent) Add(member string) {
	m.c.Add(member)
}

func (m *consistent) Adds(members ...string) {
	for idx := range members {
		m.c.Add(members[idx])
	}
}

func (m *consistent) Members() []string {
	return m.c.Members()
}

func (m *consistent) Remove(member string) {
	m.c.Remove(member)
}

func (m *consistent) Get(key *string) (string, error) {
	return m.c.Get(*key)
}

func (m *consistent) Clear() {
	m.c.Set(nil)
}
