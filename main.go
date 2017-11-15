package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "A deadman's switch for Prometheus Alertmanager compatible notifications.")
	app.HelpFlag.Short('h')

	var amURL string
	var interval model.Duration
	app.Flag("am.url", "Alertmanager URL to send alerts to.").Default("http://localhost:9093/api/v1/alerts").StringVar(&amURL)
	app.Flag("deadman.interval", "The heartbeat interval. An alert is sent if no heartbeat is sent.").Default("30s").SetValue(&interval)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	pinger := make(chan time.Time)
	go http.ListenAndServe(":9095", simpleHandler(pinger))

	d, err := NewDeadMan(pinger, time.Duration(interval), amURL)
	if err != nil {
		log.Fatalln(err)
	}

	d.Run()
}

func simpleHandler(pinger chan<- time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pinger <- time.Now()
		fmt.Fprint(w, "")
	}
}
