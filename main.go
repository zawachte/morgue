package main

import (
	"context"
	"net"
	"net/http"
	"syscall"
	"time"

	"flag"

	"github.com/spf13/pflag"
	"github.com/zawachte/morgue/internal/runner"
)

func main() {
	var retention time.Duration
	var backupFrequency time.Duration
	var metricsScrapeFrequency time.Duration
	var unixSocket string
	var backupPath string
	var telegrafLocation string
	var influxDLocation string

	fs := pflag.CommandLine
	fs.DurationVar(&retention,
		"retention",
		time.Hour,
		"retention time for stored metrics",
	)
	fs.DurationVar(&backupFrequency,
		"backup-frequency",
		time.Hour,
		"period for creating database backups",
	)
	fs.DurationVar(&metricsScrapeFrequency,
		"metrics-scrape-frequency",
		time.Second*20,
		"period for creating database backups",
	)
	fs.StringVar(&unixSocket,
		"unix-socket",
		"morgue.sock",
		"unix socket path",
	)

	fs.StringVar(&backupPath,
		"backup-path",
		".",
		"path for database backups",
	)

	fs.StringVar(&telegrafLocation,
		"telegraf-location",
		"/usr/local/bin/telegraf",
		"location of the telegraf binary",
	)
	fs.StringVar(&influxDLocation,
		"influxd-location",
		"/usr/local/bin/influxd",
		"location of the influxd binary",
	)

	fs.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	run, err := runner.NewRunner(runner.RunnerParams{
		BackupFrequency:  backupFrequency,
		BackupPath:       backupPath,
		Retention:        retention,
		InfluxDLocation:  influxDLocation,
		TelegrafLocation: telegrafLocation,
	})
	if err != nil {
		panic(err)
	}

	err = run.Run(context.Background())
	if err != nil {
		panic(err)
	}

	syscall.Unlink(unixSocket)

	// remove when we get there
	unixListener, err := net.Listen("unix", unixSocket)
	if err != nil {
		panic(err)
	}
	defer unixListener.Close()

	server := http.Server{}
	server.Serve(unixListener)
}
