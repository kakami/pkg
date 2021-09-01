package types_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

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

func Test_Add(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// var dp int = 120
	// var idx int = 20
	// for i := 20; i < 2000; i++ {
	// 	dc := testCC(i)
	// 	if dp > dc || dc < 7 {
	// 		dp = dc
	// 		idx = i
	// 	}
	// }
	// fmt.Println("use", idx, testCC(idx))
	idx := 1237
	fmt.Println("use", idx, testCC(idx))
}

var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")

func randomString(length int) string {
	var out []byte
	for i := 0; i < length; i++ {
		out = append(out, chars[rand.Int()%len(chars)])
	}
	return string(out)
}

func testCC(num int) int {
	cc := types.NewConsistent()
	cc.NumberOfReplicas = num
	cc.UseFnv = true

	cc.Add("42.240.152.34")
	cc.Add("42.240.152.35")
	cc.Add("42.240.152.36")
	cc.Add("42.240.152.195")
	cc.Add("42.240.152.196")
	// cc.Add(string(net.ParseIP("42.240.152.34")))
	// cc.Add(string(net.ParseIP("42.240.152.35")))
	// cc.Add(string(net.ParseIP("42.240.152.36")))
	// cc.Add(string(net.ParseIP("42.240.152.195")))
	// cc.Add(string(net.ParseIP("42.240.152.196")))

	// for i := 0; i < 5 ; i++ {
	// 	cc.Add(randomString(14))
	// }

	dp := cc.Print()
	fmt.Printf("====> %d\n", dp)

	// p, _ := cc.GetN("aaa", 110)
	// fmt.Println(p)
	return int(math.Abs(float64(dp - 100)))
}
