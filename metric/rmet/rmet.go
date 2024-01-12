package rmet

import (
    "container/ring"
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/shirou/gopsutil/v3/process"
    "go.uber.org/atomic"
)

type RMet struct {
    bytesSent  *atomic.Int64
    dataSent   *atomic.Int64
    bytesRecv  *atomic.Int64
    dataRecv   *atomic.Int64
    timestamp  time.Time
    sts        time.Time
    srates     *ring.Ring
    sbandwidth *ring.Ring
    rrates     *ring.Ring
    rbandwidth *ring.Ring
    cpus       *ring.Ring

    totalBytesSent, totalBytesRecv int64
    TotalDataSent, totalDataRecv   int64

    tt time.Time
    mu sync.Mutex
}

var (
    _proc, _ = process.NewProcess(int32(os.Getpid()))
)

func New(size int) *RMet {
    n := min(size, 1000)
    n = max(n, 10)
    rm := &RMet{
        bytesSent:  atomic.NewInt64(0),
        dataSent:   atomic.NewInt64(0),
        bytesRecv:  atomic.NewInt64(0),
        dataRecv:   atomic.NewInt64(0),
        srates:     ring.New(n),
        sbandwidth: ring.New(n),
        rrates:     ring.New(n),
        rbandwidth: ring.New(n),
        cpus:       ring.New(n),
        timestamp:  time.Now(),
        sts:        time.Now(),
    }
    return rm
}

func (r *RMet) AddBytesSent(n int64) {
    r.bytesSent.Add(n)
}

func (r *RMet) AddDataSent(n int64) {
    r.dataSent.Add(n)
}

func (r *RMet) AddBytesRecv(n int64) {
    r.bytesRecv.Add(n)
}

func (r *RMet) AddDataRecv(n int64) {
    r.dataRecv.Add(n)
}

func (r *RMet) Metrics() ([]float64, []float64, []float64, []float64, []float64) {
    r.mu.Lock()
    defer r.mu.Unlock()
    var srates, sbandwidth, rrates, rbandwidth, cpus []float64
    r.srates.Do(func(x any) {
        if x != nil {
            srates = append(srates, x.(float64))
        }
    })
    r.sbandwidth.Do(func(x any) {
        if x != nil {
            sbandwidth = append(sbandwidth, x.(float64))
        }
    })
    r.rrates.Do(func(x any) {
        if x != nil {
            rrates = append(rrates, x.(float64))
        }
    })
    r.rbandwidth.Do(func(x any) {
        if x != nil {
            rbandwidth = append(rbandwidth, x.(float64))
        }
    })
    r.cpus.Do(func(x any) {
        if x != nil {
            cpus = append(cpus, x.(float64))
        }
    })
    return srates, sbandwidth, rrates, rbandwidth, cpus
}

func (r *RMet) Tick() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.tt = r.timestamp
    r.timestamp = time.Now()
    totalBytesSent := r.bytesSent.Load()
    totalDataSent := r.dataSent.Load()
    totalBytesRecv := r.bytesRecv.Load()
    totalDataRecv := r.dataRecv.Load()
    ttd := int64(r.timestamp.Sub(r.tt) / time.Millisecond)
    if ttd < 1 {
        ttd = 1
    }
    tst := int64(r.timestamp.Sub(r.sts) / time.Millisecond)
    if tst < 1 {
        tst = 1
    }
    p, _ := _proc.CPUPercent()
    out := fmt.Sprintf(`sent rate(c/a): %s, %s, bandwidth(c/a): %.3f, %.3f,
recv rate(c/a): %s, %s, bandwidth(c/a): %.3f, %.3f,
cpu: %.3f%%, data sent: %d, recv: %d`,
        rate(totalDataSent-r.TotalDataSent, ttd), rate(totalDataSent, tst),
        float64(totalBytesSent-r.totalBytesSent+1)/float64(totalDataSent-r.TotalDataSent+1),
        float64(totalBytesSent+1)/float64(totalDataSent+1),
        rate(totalDataRecv-r.totalDataRecv, ttd), rate(totalDataRecv, tst),
        float64(totalBytesRecv-r.totalBytesRecv+1)/float64(totalDataRecv-r.totalDataRecv+1),
        float64(totalBytesRecv+1)/float64(totalDataRecv+1),
        p,
        totalDataSent,
        totalDataRecv,
    )
    r.sbandwidth = r.sbandwidth.Next()
    r.sbandwidth.Value = float64(totalBytesSent-r.totalBytesSent+1) / float64(totalDataSent-r.TotalDataSent+1)
    r.srates = r.srates.Next()
    r.srates.Value = float64(totalDataSent-r.TotalDataSent) / float64(ttd)
    r.rbandwidth = r.rbandwidth.Next()
    r.rbandwidth.Value = float64(totalBytesRecv-r.totalBytesRecv+1) / float64(totalDataRecv-r.totalDataRecv+1)
    r.rrates = r.rrates.Next()
    r.rrates.Value = float64(totalDataRecv-r.totalDataRecv) / float64(ttd)
    r.cpus = r.cpus.Next()
    r.cpus.Value = p
    r.totalBytesSent = totalBytesSent
    r.totalBytesRecv = totalBytesRecv
    r.TotalDataSent = totalDataSent
    r.totalDataRecv = totalDataRecv
    return out
}

func rate(n int64, d int64) string {
    if n > 3*1024*1024 {
        return fmt.Sprintf("%.3fMB/s", float64(n/1024)/float64(d))
    }
    return fmt.Sprintf("%.3fKB/s", float64(n)/float64(d))
}
