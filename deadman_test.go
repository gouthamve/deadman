package main

import (
	"testing"
	"time"
)

func TestDeadManDoesntTrigger(t *testing.T) {
	pinger := time.NewTicker(10 * time.Millisecond)
	defer pinger.Stop()

	called := false

	d := newDeadMan(pinger.C, 20*time.Millisecond, func() error {
		called = true
		return nil
	})

	go d.Run()
	defer d.Stop()

	time.Sleep(100 * time.Millisecond)
	if called == true {
		t.Fatal("deadman triggered!")
	}
}

func TestDeadManTriggers(t *testing.T) {
	pinger := time.NewTicker(30 * time.Millisecond)
	defer pinger.Stop()

	called := false

	d := newDeadMan(pinger.C, 20*time.Millisecond, func() error {
		called = true
		return nil
	})

	go d.Run()
	defer d.Stop()

	time.Sleep(100 * time.Millisecond)
	if called == false {
		t.Fatal("deadman did not trigger!")
	}
}
