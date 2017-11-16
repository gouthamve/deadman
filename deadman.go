package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var (
	ticksTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_ticks_total",
			Help: "The total ticks passed in this snitch",
		},
	)

	ticksNotified = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_ticks_notified",
			Help: "The number of ticks where notifications were sent.",
		},
	)

	failedNotifications = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deadman_notifications_failed",
			Help: "The number of failed notifications.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		ticksTotal,
		ticksNotified,
		failedNotifications,
	)
}

func NewDeadMan(pinger <-chan time.Time, interval time.Duration, amURL string, logger log.Logger) (*Deadman, error) {
	return newDeadMan(pinger, interval, amNotifier(amURL), logger), nil
}

type Deadman struct {
	pinger   <-chan time.Time
	interval time.Duration
	ticker   *time.Ticker
	closer   chan struct{}

	notifier func() error

	logger log.Logger
}

func newDeadMan(pinger <-chan time.Time, interval time.Duration, notifier func() error, logger log.Logger) *Deadman {
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
			ticksTotal.Inc()

			if !skip {
				ticksNotified.Inc()
				if err := d.notifier(); err != nil {
					failedNotifications.Inc()
					level.Error(d.logger).Log("err", err)
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
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
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
