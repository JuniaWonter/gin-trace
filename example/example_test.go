package main

import (
	"context"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*6)
	defer cancel()
	ch := make(chan interface{}, 1)
	call := func() {
		ch <- <-time.After(time.Second * 10)
	}

	go call()

	select {
	case <-ch:
		t.Log("OK")
	case <-ctx.Done():
		t.Log("timeout")
	case <-time.After(time.Second * 10):
		t.Log("timeafter")
	}
}
