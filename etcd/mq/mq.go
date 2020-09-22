package mq

import (
	"context"
	"sync"

	v3 "github.com/coreos/etcd/clientv3"
)

const (
	mqMark = "e3a165ab0e8e"
)

type Consumer struct {
	cli  *v3.Client
	mark string

	subs map[string]*subscriber
	mu   sync.Mutex
	once sync.Once
}

func NewConsumer(cli *v3.Client, mark string) *Consumer {
	c := &Consumer{
		cli:  cli,
		mark: mark,
		subs: make(map[string]*subscriber),
	}

	return c
}

func (c *Consumer) Subscribe(ctx context.Context, key string) v3.WatchChan {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.subs[key]; ok {
		ch := make(chan v3.WatchResponse)
		close(ch)
		return ch
	}

	s := newSubscriber(ctx, c.cli, key)
	c.subs[key] = s
	return s.watch()
}

func (c *Consumer) Retract(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.subs[key]; ok {
		s.retract()
	}
}
