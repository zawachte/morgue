package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"flag"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/zawachte/morgue/internal/runner"
	"go.uber.org/zap"
)

func main() {
	var retention time.Duration
	var backupFrequency time.Duration
	var metricsScrapeFrequency time.Duration
	var unixSocket string
	var backupPath string
	var telegrafLocation string
	var influxDLocation string
	var storageDriver string
	var awsRegion string
	var awsBucketName string
	var serviceMode bool

	fs := pflag.CommandLine
	fs.BoolVar(&serviceMode,
		"service-mode",
		false,
		"switch to configure telegraf and influxd as systemd services",
	)
	fs.DurationVar(&retention,
		"retention",
		6*time.Hour,
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
	fs.StringVar(&storageDriver,
		"storage-driver",
		"local",
		"type of storage driver [local, aws]",
	)
	fs.StringVar(&awsRegion,
		"aws-region",
		"us-east-1",
		"aws region",
	)
	fs.StringVar(&awsBucketName,
		"aws-s3-bucket",
		"",
		"name of the s3 bucket",
	)

	fs.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}

	runnerParams := runner.RunnerParams{
		BackupFrequency:  backupFrequency,
		BackupPath:       backupPath,
		Retention:        retention,
		InfluxDLocation:  influxDLocation,
		TelegrafLocation: telegrafLocation,
		Logger:           *logger,
		ServiceMode:      serviceMode,
	}

	if storageDriver == "aws" {
		runnerParams.AWSParams = &runner.AWSParams{
			Region:       awsRegion,
			S3BucketName: awsBucketName,
		}
	}

	run, err := runner.NewRunner(runnerParams)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	err = run.Run(context.Background())
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// TODO add metrics
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
