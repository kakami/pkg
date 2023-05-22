package alerter

import (
	"errors"
	"sync"
	"time"
)

var (
	Mark    = "0.0.0.0"
	Webhook = ""
)

var errPool struct {
	ech chan error
	fch chan error
}

func init() {
	errPool.ech = make(chan error, 10)
	errPool.fch = make(chan error, 10)

	go start()
}

func start() {
	mu := &sync.Mutex{}
	var (
		eErr, fErr         error
		eOccured, fOccured int
	)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if eOccured > 0 {
				SendAlert("-------- WARNING --------", eOccured, eErr.Error())
				eErr = nil
				eOccured = 0
			}
			if fOccured > 0 {
				SendAlert("-------- FATAL --------", fOccured, fErr.Error())
				fErr = nil
				fOccured = 0
			}
			mu.Unlock()
		case err := <-errPool.ech:
			mu.Lock()
			eOccured++
			eErr = errors.Join(eErr, err)
			mu.Unlock()
		case err := <-errPool.fch:
			mu.Lock()
			fOccured++
			fErr = errors.Join(fErr, err)
			mu.Unlock()
		}
	}
}
