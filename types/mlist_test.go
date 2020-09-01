package types_test

import (
	"math/rand"
	"testing"
	"time"

	"pkg/types"
)

var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")

func randomString(length int) string {
	var out []byte
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		out = append(out, chars[rand.Int()%len(chars)])
	}
	return string(out)
}

func Test_List(t *testing.T) {
	list := types.NewList()

	var ss5 string
	for i := 0; i < 10; i++ {
		ss := randomString(10)
		list.PushBack(ss, ss)
		if i == 5 {
			ss5 = ss
		}
	}

	var front string
	for {
		ss := list.Front().Value.(string)
		if front == "" {
			front = ss
		} else if front == ss {
			break
		}
		t.Log(ss)
		list.MoveToBack(list.Front())
	}

	sss := randomString(15)
	e := list.Find(ss5)
	if e == nil {
		t.Error(">>> Find")
		return
	}
	t.Log("========= InsertAfter " + ss5 + " ========")
	list.InsertAfter(sss, sss, e)
	t.Log(e.Next().Value.(string))
	for e := list.Front(); e != nil; e = e.Next() {
		t.Log(e.Value.(string))
	}

	t.Log("========= InsertBefore " + ss5 + " ========")
	list.InsertBefore(sss, sss, e)
	t.Log(e.Next().Value.(string))
	for e := list.Front(); e != nil; e = e.Next() {
		t.Log(e.Value.(string))
	}

	t.Log("========= PushFront ========")
	list.PushFront(sss, sss)
	e = list.Front()
	if e == nil {
		t.Error(">>> PushFront")
		return
	}
	for e := list.Front(); e != nil; e = e.Next() {
		t.Log(e.Value.(string))
	}

	t.Log("========= PushBack ========")
	list.PushBack(sss, sss)
	e = list.Back()
	if e == nil {
		t.Error(">>> PushBack")
		return
	}
	for e := list.Front(); e != nil; e = e.Next() {
		t.Log(e.Value.(string))
	}

	t.Log("========= Remove " + sss + " ========")
	list.RemoveByKey(sss)
	for e := list.Front(); e != nil; e = e.Next() {
		t.Log(e.Value.(string))
	}

	t.Log("========= Remove ========")
	for {
		e := list.Front()
		if e == nil {
			break
		}
		t.Log(e.Value.(string))
		list.Remove(e)
	}
}
