package runner

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/zawachte/morgue/internal/servicemanager"
	"github.com/zawachte/morgue/internal/storagedriver"
	"github.com/zawachte/morgue/pkg/influx"
	"github.com/zawachte/morgue/pkg/influx_cli"
	"github.com/zawachte/morgue/pkg/influxd"
)

type Runner interface {
	Run(context.Context) error
}

type runner struct {
	retention       time.Duration
	backupFrequency time.Duration
	storageDriver   storagedriver.StorageDriver
	svcManager      servicemanager.ServiceManager
}

type RunnerParams struct {
	Retention        time.Duration
	BackupFrequency  time.Duration
	BackupPath       string
	InfluxDLocation  string
	TelegrafLocation string
}

func NewRunner(params RunnerParams) (Runner, error) {
	sd, err := storagedriver.NewStorageDriver(storagedriver.StorageDriverParams{
		LocalStorageLocation: params.BackupPath,
	})
	if err != nil {
		return nil, err
	}

	svcm := servicemanager.NewServiceManager(false, servicemanager.ServiceManagerParams{
		InfluxDLocation:  params.InfluxDLocation,
		TelegrafLocation: params.TelegrafLocation,
	})

	return &runner{
		retention:       params.Retention,
		backupFrequency: params.BackupFrequency,
		storageDriver:   sd,
		svcManager:      svcm,
	}, nil
}

func (r *runner) setupInflux(token, password string) error {
	influxCli, err := influx_cli.NewClient()
	if err != nil {
		return err
	}

	params := influx_cli.SetupInfluxParams{
		Username:  influx.DefaultUsername,
		Password:  password,
		AuthToken: token,
		Org:       influx.DefaultOrgName,
		Bucket:    influx.DefaultBucketName,
		Retention: r.retention.String(),
	}

	err = influxCli.SetupInflux(params)
	if err != nil {
		return err
	}
	return nil
}

func (r *runner) Run(ctx context.Context) error {
	err := r.svcManager.RunInfluxD()
	if err != nil {
		return nil
	}

	influxd.WaitForInfluxDReady()

	token := generateToken()
	password := generateToken()
	err = r.setupInflux(token, password)
	if err != nil {
		return err
	}

	err = r.svcManager.RunTelegraf(token)
	if err != nil {
		return nil
	}

	err = r.runBackupAndStore()
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) runBackupAndStore() error {

	influxCli, err := influx_cli.NewClient()
	if err != nil {
		return err
	}

	backupParams := influx_cli.BackupInfluxParams{
		Org:    influx.DefaultOrgName,
		Bucket: influx.DefaultBucketName,
		Path:   r.storageDriver.GetLocalStorageLocation(),
	}

	go func() {
		for {
			time.Sleep(r.backupFrequency)
			err = influxCli.BackupInflux(backupParams)
			if err != nil {
				panic(err)
			}

			err = r.storageDriver.UploadTar(r.storageDriver.GetLocalStorageLocation())
			if err != nil {
				panic(err)
			}
		}
	}()

	return nil
}

func generateToken() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	return b.String()
}
