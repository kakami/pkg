package aqc

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"

	"pkg/types"
)

const (
	acMark = "3080ed7b"
)

const (
	acField = iota
	keyField
	memberField
	timeField
	fieldCount
)

type Handler interface {
	HandleKey(string)
	UnderRepair() <-chan struct{}
	Repaired() <-chan struct{}
}

type Logger interface {
	Debug(interface{}, ...interface{})
	Info(interface{}, ...interface{})
	Error(interface{}, ...interface{})
}

type AcQueue struct {
	Member  types.Membership
	mLocker sync.Mutex

	logger Logger
	ctx    context.Context
	cli    *v3.Client
	mark   string

	workers  map[string]*worker
	prevKeys types.Set
}

func NewAcQueue(ctx context.Context, cli *v3.Client, logger Logger) *AcQueue {
	q := &AcQueue{
		Member:   types.NewRoundRobinMembership(),
		ctx:      ctx,
		cli:      cli,
		logger:   logger,
		workers:  make(map[string]*worker),
		prevKeys: types.NewUnsafeSet(),
	}
	if q.ctx == nil {
		q.ctx = cli.Ctx()
	}
	return q
}

func (ac *AcQueue) RegisterMember(member string, handler Handler, interval int64) error {
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	if _, ok := ac.workers[member]; ok {
		return errors.New("AcQueue::RegieterMember member existed")
	}
	worker := newWorker(ac, member, handler, interval)
	ac.workers[member] = worker

	return nil
}

func (ac *AcQueue) UnregisterMember(member string) error {
	if err := ac.retract(member); err != nil {
		return errors.WithMessagef(err, "AcQueue::UnregisterMember retract %s", member)
	}
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	if worker, ok := ac.workers[member]; ok {
		worker.cancel()
		delete(ac.workers, member)
	}
	return nil
}

func (ac *AcQueue) PreAdd(key string) {
	ac.prevKeys.Add(key)
}

func (ac *AcQueue) Init(mark string) error {
	if len(ac.workers) < 1 {
		return errors.New("AcQueue::Init No worker registered")
	}
	ctx, cancel := context.WithTimeout(ac.ctx, 3*time.Second)
	defer cancel()
	resp, err := ac.cli.Get(ctx, filepath.Join(acMark, mark), v3.WithPrefix())
	if err != nil {
		return errors.WithMessage(err, "AcQueue::Init etcd Get")
	}
	for _, kvs := range resp.Kvs {
		parts := strings.Split(string(kvs.Key), "/")
		if len(parts) != fieldCount {
			continue
		}
		if ac.prevKeys.Contains(parts[keyField]) {
			if worker, ok := ac.workers[parts[memberField]]; ok {
				ac.prevKeys.Remove(parts[keyField])
				worker.add(parts[keyField])
			}
		} else {
			if _, err = ac.cli.Delete(ctx, string(kvs.Key)); err != nil {
				return errors.WithMessage(err, "AcQueue::Init delete dead key")
			}
		}
	}

	left := ac.prevKeys.Values()
	var member string
	for idx := range left {
		member, err = ac.Member.Get(&left[idx])
		if err != nil {
			return errors.WithMessage(err, "AcQueue::Init dispatch work")
		}
		if worker, ok := ac.workers[member]; ok {
			worker.add(left[idx])
		} else {
			return errors.New("!!!!! this should be unreachable")
		}
	}

	for _, worker := range ac.workers {
		go worker.run()
	}
	return nil
}

func (ac *AcQueue) Dispatch(key string) error {
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	return ac.dispatch(key)
}

func (ac *AcQueue) dispatch(key string) error {
	member, err := ac.Member.Get(&key)
	if err != nil {
		return errors.WithMessage(err, "AcQueue::Add Member Get")
	}
	if worker, ok := ac.workers[member]; ok {
		worker.add(key)
		ctx, cancel := context.WithTimeout(ac.ctx, 3*time.Second)
		defer cancel()
		if _, err = ac.cli.Put(ctx, filepath.Join(acMark, ac.mark, key, member, strconv.FormatInt(time.Now().Unix(), 10)), ""); err != nil {
			return errors.WithMessage(err, "AcQueue::Add etcd put")
		}
	} else {
		return errors.New("!!!!! AcQueue Add::this should be unreachable")
	}
	return nil
}

func (ac *AcQueue) Remove(key string) error {
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	for _, worker := range ac.workers {
		worker.remove(key)
	}
	ctx, cancel := context.WithTimeout(ac.ctx, 3*time.Second)
	defer cancel()
	if _, err := ac.cli.Delete(ctx, filepath.Join(acMark, ac.mark, key)+"/", v3.WithPrefix()); err != nil {
		return errors.WithMessage(err, "AcQueue::Remove etcd delete")
	}
	return nil
}

func (ac *AcQueue) retract(member string) error {
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	ac.logger.Info("\n====================== AcQueue::retract =========================")
	ac.logger.Info(">> retract [%s], something wrong!!!", member)
	ac.Member.Remove(member)

	if worker, ok := ac.workers[member]; ok {
		tasks := worker.tasks()
		for idx := range tasks {
			if err := ac.dispatch(tasks[idx]); err != nil {
				return errors.WithMessagef(err, "AcQueue::retract dispatch [%s] err: %s", tasks[idx], err.Error())
			}
		}
	}

	return nil
}

func (ac *AcQueue) activate(member string) error {
	ac.mLocker.Lock()
	defer ac.mLocker.Unlock()
	if _, ok := ac.workers[member]; !ok {
		return errors.Errorf("worker [%s] not registered")
	}
	ac.Member.Add(member)
	return nil
}
