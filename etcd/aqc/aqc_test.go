package aqc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	v3 "github.com/coreos/etcd/clientv3"

	"pkg/etcd/aqc"
)

func Test_AcQueue(t *testing.T) {
	ll.t = t
	ctx := context.Background()
	cli, err := v3.New(v3.Config{
		Endpoints:   []string{"10.68.192.112:2379", "10.68.192.113:2379"},
		DialTimeout: 5 * time.Second,
		Context:     ctx,
	})
	if err != nil {
		t.Error(err)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	if _, err = cli.Get(cctx, "check_server_avaliable"); err != nil {
		t.Error(err)
		return
	}
	cancel()

	ac := aqc.NewAcQueue(ctx, cli, &ll)
	virtualPreAdd(ac)
	registerMember(ac)
	if err = ac.Init("actest"); err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Minute)
}

func virtualPreAdd(ac *aqc.AcQueue) {
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%d.acqueue.com/app%d/name%d", i, i, i)
		ac.PreAdd(key)
	}
}

func registerMember(ac *aqc.AcQueue) {
	for i := 0; i < 3; i++ {
		member := fmt.Sprintf("handler_%d", i)
		h := newHandler(member)
		ac.RegisterMember(member, h, 3)
	}
}
