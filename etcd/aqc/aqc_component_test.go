package aqc_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

var ll logger

type logger struct {
	t *testing.T
}

func (l *logger) Debug(arg0 interface{}, args ...interface{}) {
	var msg string
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		msg = fmt.Sprintf(first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		msg = first()
	default:
		// Build a format string so that it will be similar to Sprint
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	l.t.Log(msg)
}

func (l *logger) Info(arg0 interface{}, args ...interface{}) {
	var msg string
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		msg = fmt.Sprintf(first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		msg = first()
	default:
		// Build a format string so that it will be similar to Sprint
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	l.t.Log(msg)
}

func (l *logger) Error(arg0 interface{}, args ...interface{}) error {
	var msg string
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		msg = fmt.Sprintf(first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		msg = first()
	default:
		// Build a format string so that it will be similar to Sprint
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	l.t.Log(msg)
	return errors.New(msg)
}

// //////////////////////////////////////////////////////////////////////////
type handler struct {
	id   string
	urtc chan struct{}
	rtc  chan struct{}

	underRepair bool
	output      bool
}

func newHandler(id string) *handler {
	h := &handler{
		id:   id,
		urtc: make(chan struct{}),
		rtc:  make(chan struct{}),
	}
	go h.broken()
	return h
}

func (h *handler) HandleKey(task string) {
	fmt.Println(">>>", h.id, task)
}

func (h *handler) UnderRepair() <-chan struct{} {
	return h.urtc
}

func (h *handler) Repaired() <-chan struct{} {
	return h.rtc
}

func (h *handler) broken() {
	time.Sleep(3 * time.Second)
	if h.id == "handler_0" {
		h.urtc <- struct{}{}
	}
	time.Sleep(2 * time.Second)
	if h.id == "handler_0" {
		h.rtc <- struct{}{}
	}
	time.Sleep(time.Second)
	if h.id == "handler_1" {
		h.urtc <- struct{}{}
	}
	time.Sleep(time.Second)
	if h.id == "handler_1" {
		h.rtc <- struct{}{}
	}
	if h.id == "handler_2" {
		h.urtc <- struct{}{}
	}
}
