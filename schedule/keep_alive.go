package schedule

import (
    "context"
    "sort"
    "sync"
    "time"
)

type ExpireFunc func(Cell)

type Cell interface {
    Key() string
    IsExpired() bool
    Less(Cell) bool
}

type KeepAlive struct {
    ctx      context.Context
    cancel   context.CancelFunc
    interval time.Duration
    vmap     map[string]Cell
    vs       []Cell
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

func (k *KeepAlive) Add(cell Cell) {
    k.mu.Lock()
    defer k.mu.Unlock()

    k.vmap[cell.Key()] = cell
    cells := make([]Cell, 0, len(k.vmap))
    for _, v := range k.vmap {
        cells = append(cells, v)
    }
    k.vs = cells
    sort.Sort(k)
}

//////////////////////////////
/// sort

func (k *KeepAlive) Swap(i, j int) {
    k.vs[i], k.vs[j] = k.vs[j], k.vs[i]
}

func (k *KeepAlive) Less(i, j int) bool {
    return k.vs[i].Less(k.vs[j])
}

func (k *KeepAlive) Len() int {
    return len(k.vs)
}

func (k *KeepAlive) checkAlive() int {
    k.mu.Lock()
    defer k.mu.Unlock()

    sort.Sort(k)

    var cnt int
    for idx := range k.vs {
        if !k.vs[idx].IsExpired() {
            break
        }
        if k.exf != nil {
            k.exf(k.vs[idx])
        }
        delete(k.vmap, k.vs[idx].Key())
        cnt++
    }
    k.vs = k.vs[cnt:]
    return cnt
}
