package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/common/model"
)

func NewDeadMan(pinger <-chan time.Time, interval time.Duration, amURL string) (*Deadman, error) {
	return newDeadMan(pinger, interval, amNotifier(amURL)), nil
}

type Deadman struct {
	pinger   <-chan time.Time
	interval time.Duration
	ticker   *time.Ticker
	closer   chan struct{}

	notifier func() error
}

func newDeadMan(pinger <-chan time.Time, interval time.Duration, notifier func() error) *Deadman {
	return &Deadman{
		pinger:   pinger,
		interval: interval,
		notifier: notifier,
		closer:   make(chan struct{}),
	}
}

func (d *Deadman) Run() error {
	d.ticker = time.NewTicker(d.interval)

	skip := false

	for {
		select {
		case <-d.ticker.C:
			if !skip {
				// TODO: Handle errors.
				if err := d.notifier(); err != nil {
					fmt.Println(err)
				}
			}
			skip = false

		case <-d.pinger:
			skip = true

		case <-d.closer:
			break
		}
	}

	return nil
}

func (d *Deadman) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}

	d.closer <- struct{}{}
}

func amNotifier(amURL string) func() error {
	alerts := []*model.Alert{{
		Labels: model.LabelSet{
			model.LabelName("alertname"): model.LabelValue("DeadmanDead"),
		},
	}}

	b, err := json.Marshal(alerts)
	if err != nil {
		log.Fatalln(err)
	}

	return func() error {
		client := &http.Client{}
		resp, err := client.Post(amURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("bad response status %v", resp.Status)
		}

		return nil
	}
}
