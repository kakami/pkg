package schedule

import (
    "context"
    "sort"
    "sync"
    "time"
)

type ExpireFunc func(string)

type cell struct {
    key string
    ttl int64
}

type KeepAlive struct {
    ctx      context.Context
    cancel   context.CancelFunc
    interval time.Duration
    vmap     map[string]*cell
    vs       []*cell
    exf      ExpireFunc
    mu       sync.RWMutex
}

func NewKeepAlive(d time.Duration, f ExpireFunc) *KeepAlive {
    k := &KeepAlive{
        interval: d,
        exf:      f,
    }
    if k.interval < time.Second {
        k.interval = time.Second
    }
    return k
}

func (k *KeepAlive) Start(ctx context.Context) {
    k.ctx, k.cancel = context.WithCancel(ctx)
    defer k.cancel()
    ticker := time.NewTicker(k.interval)
    select {
    case <-k.ctx.Done():
        return
    case <-ticker.C:
        k.checkAlive()
    }
}

func (k *KeepAlive) Stop() {
    if k.cancel != nil {
        k.cancel()
    }
}

func (k *KeepAlive) IsAlive(key string) bool {
    k.mu.RLock()
    defer k.mu.RUnlock()
    _, ok := k.vmap[key]
    return ok
}

func (k *KeepAlive) Set(key string, ttl int64) {
    k.mu.Lock()
    defer k.mu.Unlock()

    if v, ok := k.vmap[key]; !ok {
        v = &cell{
            key: key,
            ttl: ttl,
        }
        k.vmap[key] = v
    } else {
        v.ttl = ttl
    }
}

//////////////////////////////
/// sort

func (k *KeepAlive) Swap(i, j int) {
    k.vs[i], k.vs[j] = k.vs[j], k.vs[i]
}

func (k *KeepAlive) Less(i, j int) bool {
    return k.vs[i].ttl < k.vs[j].ttl
}

func (k *KeepAlive) Len() int {
    return len(k.vs)
}

func (k *KeepAlive) checkAlive() int {
    k.mu.Lock()
    defer k.mu.Unlock()

    sort.Sort(k)
    tn := time.Now().Unix()

    var cnt int
    for idx := range k.vs {
        if k.vs[idx].ttl > tn {
            break
        }
        if k.exf != nil {
            k.exf(k.vs[idx].key)
        }
        delete(k.vmap, k.vs[idx].key)
        cnt++
    }
    return cnt
}
