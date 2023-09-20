package alerter

import (
    "context"
    "errors"
    "fmt"
    "sync"
    "time"
)

var (
    _ctx      context.Context
    _cancel   context.CancelFunc
    _alerters []*Alerter
    _once     sync.Once
    _mu       sync.Mutex
)

var (
    _ErrFatal = fmt.Errorf("this is a fatal error")
)

type HandleFunc func(title string, num int, msg string)

func Start(ctx context.Context) error {
    initialized := true
    _once.Do(func() {
        initialized = false
    })
    if initialized {
        return fmt.Errorf("alerter already started")
    }
    _ctx, _cancel = context.WithCancel(ctx)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-_ctx.Done():
            return context.Canceled
        case <-ticker.C:
            _mu.Lock()
            alts := _alerters
            _mu.Unlock()
            for _, a := range alts {
                a.SendAlert()
            }
        }
    }
}

func Stop() {
    if f := _cancel; f != nil {
        f()
    }
}

type Alerter struct {
    title string
    err   error
    total int
    f     HandleFunc
    mu    sync.Mutex
}

func New(title string, f HandleFunc) *Alerter {
    alert := &Alerter{
        title: title,
        err:   _ErrFatal,
        f:     f,
    }
    _mu.Lock()
    defer _mu.Unlock()
    _alerters = append(_alerters, alert)
    return alert
}

func (a *Alerter) Fatal(err error) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.total++
    if a.total < 10 {
        a.err = errors.Join(a.err, err)
    }
    return err
}

func (a *Alerter) SendAlert() {
    a.mu.Lock()
    if a.total == 0 {
        a.mu.Unlock()
        return
    }
    num := a.total
    err := a.err
    a.err = _ErrFatal
    a.total = 0
    a.mu.Unlock()

    if a.f != nil {
        a.f(a.title, num, err.Error())
    }
}
