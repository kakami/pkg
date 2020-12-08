package types_test

import (
	"fmt"
	"testing"

	"github.com/kakami/pkg/types"
)

func Benchmark_ConsistentGet(b *testing.B) {
	c := types.NewConsistent()
	for i := 0; i < 10; i++ {
		c.Add(fmt.Sprintf("node_%d", i))
	}

	var keys []string
	for i := 0; i < 100; i++ {
		keys = append(keys, fmt.Sprintf("key_%d", i))
	}

	length := len(keys)
	for i := 0; i < b.N; i++ {
		c.Get(keys[i%length])
	}
}

type node struct {
	id  string
	cnt int
}

func (n *node) String() string {
	return n.id
}

func Test_IConsistent(t *testing.T) {
	ic := types.NewIConsistent()
	for i := 0; i < 10; i++ {
		ic.Add(&node{
			id: fmt.Sprintf("node_%d", i),
		})
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		v, err := ic.Get(key)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(v.(*node).String())
		v.(*node).cnt++
	}

	ms := ic.Members()
	for idx := range ms {
		t.Log(ms[idx].(*node).String(), "===>", ms[idx].(*node).cnt)
	}
}

func Benchmark_IConsistentGet(b *testing.B) {
	ic := types.NewIConsistent()
	for i := 0; i < 10; i++ {
		ic.Add(&node{
			id: fmt.Sprintf("node_%d", i),
		})
	}

	var keys []string
	for i := 0; i < 100; i++ {
		keys = append(keys, fmt.Sprintf("key_%d", i))
	}

	length := len(keys)
	for i := 0; i < b.N; i++ {
		n, _ := ic.Get(keys[i%length])
		_ = n.(*node)
	}
}
