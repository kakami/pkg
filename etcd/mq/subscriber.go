package mq

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"pkg/util"
)

const (
	mqField = iota
	aliveField
	keyField
	memberField
	timeField
	fieldCount
)

type CloseChan <-chan struct{}

type subscriber struct {
	cli       *v3.Client
	ctx, gctx context.Context
	cancel    context.CancelFunc
	g         *errgroup.Group
	key, mark string
	watchKey  string

	closeChan chan struct{}

	// used by leader
	members Membership
	lgctx   context.Context
	lg      *errgroup.Group

	mu   sync.Mutex
	once sync.Once
}

func newSubscriber(ctx context.Context, cli *v3.Client, key string) *subscriber {
	s := &subscriber{
		cli:       cli,
		key:       key,
		mark:      util.RandomString(10),
		members:   NewRoundRobinMembership(),
		closeChan: make(chan struct{}),
	}

	s.members.Add(s.mark)
	if ctx == nil {
		ctx = cli.Ctx()
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	return s
}

func (s *subscriber) retract() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *subscriber) watch() v3.WatchChan {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.g, s.gctx = errgroup.WithContext(s.ctx)
				s.g.Go(s.keepAlive)
				s.g.Go(s.campaign)
				s.g.Wait()
			}
		}
	}()

	watcher := v3.NewWatcher(s.cli)
	s.watchKey = filepath.Join(mqMark, s.mark) + "/"
	return watcher.Watch(s.gctx, s.watchKey, v3.WithPrefix())
}

func (s *subscriber) campaign() error {
	for {
		select {
		case <-s.gctx.Done():
			return errors.New("campaign ctx done")
		default:
		}
		ss, err := concurrency.NewSession(s.cli, concurrency.WithTTL(10))
		if err != nil {
			continue
		}
		e := concurrency.NewElection(ss, "subscribe-"+s.key)
		if err = e.Campaign(s.gctx, s.mark); err != nil {
			continue
		}
		s.takeoff(ss)
		ss.Close()
	}
}

func (s *subscriber) takeoff(ss *concurrency.Session) {
	s.members.Clear()
	s.lg, s.lgctx = errgroup.WithContext(s.gctx)
	s.lg.Go(s.watchAlive)
	s.lg.Go(s.watchPublish)
	s.lg.Go(func() error {
		select {
		case <-s.gctx.Done():
			return errors.New("subscribe gctx done")
		case <-ss.Done():
			return errors.New("concurrency session done")
		}
	})
	s.lg.Wait()
}

func (s *subscriber) watchPublish() error {
	return nil
}

func (s *subscriber) watchAlive() error {
	aliveKey := filepath.Join(mqMark, "alive", s.key)
	resp, err := s.cli.Get(s.lgctx, aliveKey, v3.WithPrefix())
	if err != nil {
		return errors.WithMessage(err, "watchAlive get members")
	}
	for _, kvs := range resp.Kvs {
		parts := strings.Split(string(kvs.Key), "/")
		if len(parts) != fieldCount {
			continue
		}
		s.members.Add(parts[memberField])
	}

	watcher := v3.NewWatcher(s.cli)
	wcc := watcher.Watch(s.lgctx, aliveKey, v3.WithPrefix())
	for {
		select {
		case <-s.lgctx.Done():
			return errors.New("watchAlive ctx done")
		case resp, ok := <-wcc:
			if !ok {
				return errors.New("watchAlive closed")
			}
			s.mu.Lock()
			for _, event := range resp.Events {
				parts := strings.Split(string(event.Kv.Key), "/")
				if len(parts) == fieldCount {
					if event.IsCreate() {
						s.members.Add(parts[memberField])
					} else if event.Type == mvccpb.DELETE {
						s.members.Remove(parts[memberField])
						go s.redispatch(parts[memberField])
					}
				}
			}
			s.mu.Unlock()
			if resp.Canceled {
				return errors.WithMessage(resp.Err(), "watch alive")
			}
		}
	}
	// never reached
	// return nil
}

func (s *subscriber) dispatch(value string) error {
	mark, err := s.members.Get()
	if err != nil {
		return err
	}
	_, err = s.cli.Put(s.lgctx, filepath.Join(mqMark, mark, value), strconv.FormatInt(time.Now().Unix(), 10))
	return err
}

func (s *subscriber) redispatch(member string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// mqMark/mark
	prefix := filepath.Join(mqMark, member) + "/"
	resp, err := s.cli.Get(s.lgctx, prefix, v3.WithPrefix())
	if err != nil {
		s.cancel()
		return
	}
	for _, kvs := range resp.Kvs {
		key := string(kvs.Key)
		s.dispatch(key[len(prefix):])
	}
}

func (s *subscriber) keepAlive() error {
	aliveKey := filepath.Join(mqMark, "alive", s.key, strconv.FormatInt(time.Now().Unix(), 10))
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-s.gctx.Done():
			return errors.New("gctx done")
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(s.gctx, 3*time.Second)
			_, err := s.cli.Put(ctx, aliveKey, strconv.FormatInt(time.Now().Unix(), 10))
			cancel()
			if err == nil {
				continue
			}
			ctx, cancel = context.WithTimeout(s.gctx, 5*time.Second)
			_, err = s.cli.Put(ctx, aliveKey, strconv.FormatInt(time.Now().Unix(), 10))
			if err != nil {
				cancel()
				s.once.Do(func() {
					close(s.closeChan)
				})
				s.cancel()
				return errors.WithMessage(err, "keep alive")
			}
		}
	}
}
