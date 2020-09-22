package aqc

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"pkg/types"
)

type uTask struct {
	work string
	ttl  int64
}

type worker struct {
	ac          *AcQueue
	id          string
	g           *errgroup.Group
	ctx, gctx   context.Context
	cancel      context.CancelFunc
	works       *types.List
	handler     Handler
	interval    int64
	underRepair bool
	mu          sync.RWMutex
}

func newWorker(ac *AcQueue, id string, handler Handler, interval int64) *worker {
	w := &worker{
		ac:       ac,
		id:       id,
		works:    types.NewList(),
		handler:  handler,
		interval: interval,
	}
	w.ctx, w.cancel = context.WithCancel(ac.ctx)
	return w
}

func (w *worker) add(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.works.PushFront(key, &uTask{work: key})
}

func (w *worker) remove(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.works.RemoveByKey(key)
}

func (w *worker) run() {
	w.underRepair = false
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.g, w.gctx = errgroup.WithContext(w.ctx)
			w.g.Go(w.monitor)
			w.g.Go(w.working)

			if err := w.g.Wait(); err != nil {
				w.ac.logger.Error("worker[%s]: %s", w.id, err.Error())
			}
		}
	}
}

func (w *worker) monitor() error {
	for {
		select {
		case <-w.ctx.Done():
			return errors.New("worker::monitor ctx Done")
		case <-w.handler.UnderRepair():
			w.underRepair = true
			w.ac.retract(w.id)
		case <-w.handler.Repaired():
			w.ac.activate(w.id)
			w.underRepair = false
		}
	}
}

func (w *worker) working() error {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-w.gctx.Done():
			return errors.New("worker::working gctx Done")
		case <-ticker.C:
			if w.underRepair {
				continue
			}
			tNow := time.Now().Unix()
			for {
				w.mu.Lock()
				e := w.works.Front()
				if e == nil {
					w.mu.Unlock()
					break
				}
				if task, ok := e.Value.(*uTask); ok {
					if tNow > task.ttl {
						task.ttl = tNow + w.interval
					} else {
						w.mu.Unlock()
						break
					}
					go w.handler.HandleKey(task.work)
					w.works.MoveToBack(e)
				} else {
					w.works.Remove(e)
				}
				w.mu.Unlock()
			}
		}
	}
}

func (w *worker) tasks() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var tasks []string
	for e := w.works.Front(); e != nil; e = e.Next() {
		if t, ok := e.Value.(*uTask); ok {
			tasks = append(tasks, t.work)
		}
	}
	return tasks
}
