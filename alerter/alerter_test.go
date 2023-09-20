package alerter_test

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/kakami/pkg/alerter"
)

func Test_Alert(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()
    go alerter.Start(ctx)

    fn := func(title string, num int, msg string) {
        t.Errorf("%s: %d: %s\n", title, num, msg)
    }

    alt1 := alerter.New("test1", fn)
    alt2 := alerter.New("test2", fn)
    ticker := time.NewTicker(2 * time.Second)
    var idx int64
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            idx++
            if idx > 10 {
                return
            }
            if idx%2 == 0 {
                alt1.Fatal(fmt.Errorf("test1 error %d", idx))
            } else {
                alt2.Fatal(fmt.Errorf("test2 error %d", idx))
            }
        }
    }
}
