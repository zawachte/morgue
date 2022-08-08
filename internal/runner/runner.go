package runner

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/zawachte/morgue/internal/servicemanager"
	"github.com/zawachte/morgue/internal/storagedriver"
	"github.com/zawachte/morgue/pkg/influx"
	"github.com/zawachte/morgue/pkg/influx_cli"
	"github.com/zawachte/morgue/pkg/influxd"
	"github.com/zawachte/morgue/pkg/tarutils"
	"go.uber.org/zap"
)

type Runner interface {
	Run(context.Context) error
}

type runner struct {
	retention       time.Duration
	backupFrequency time.Duration
	storageDriver   storagedriver.StorageDriver
	svcManager      servicemanager.ServiceManager
	logger          zap.Logger
}

type AWSParams struct {
	Region       string
	S3BucketName string
}

type RunnerParams struct {
	Retention        time.Duration
	BackupFrequency  time.Duration
	ServiceMode      bool
	BackupPath       string
	InfluxDLocation  string
	TelegrafLocation string
	AWSParams        *AWSParams
	Logger           zap.Logger
}

func NewRunner(params RunnerParams) (Runner, error) {

	strgDriverParams := storagedriver.StorageDriverParams{
		LocalStorageLocation: params.BackupPath,
		Logger:               params.Logger,
	}

	if params.AWSParams != nil {
		strgDriverParams.S3StorageDriverParams = &storagedriver.S3StorageDriverParams{
			Bucket: params.AWSParams.S3BucketName,
			Region: params.AWSParams.Region,
		}
	}

	sd, err := storagedriver.NewStorageDriver(strgDriverParams)
	if err != nil {
		return nil, err
	}

	svcm := servicemanager.NewServiceManager(params.ServiceMode, servicemanager.ServiceManagerParams{
		InfluxDLocation:  params.InfluxDLocation,
		TelegrafLocation: params.TelegrafLocation,
		Logger:           params.Logger,
	})

	return &runner{
		retention:       params.Retention,
		backupFrequency: params.BackupFrequency,
		storageDriver:   sd,
		svcManager:      svcm,
		logger:          params.Logger,
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

	go func() {
		for {
			time.Sleep(r.backupFrequency)
			err := r.backupAndStore(influxCli)
			if err != nil {
				r.logger.Warn(err.Error())
			}
		}
	}()

	return nil
}

func (r *runner) backupAndStore(influxClient influx_cli.Client) error {
	backupFilenamePattern := "20060102T150405Z"
	directoryName := time.Now().UTC().Format(backupFilenamePattern)

	backupParams := influx_cli.BackupInfluxParams{
		Org:    influx.DefaultOrgName,
		Bucket: influx.DefaultBucketName,
		Path:   r.storageDriver.GetLocalStorageLocation(),
	}

	defer cleanupBackup(backupParams.Path)

	backupParams.Path = path.Join(r.storageDriver.GetLocalStorageLocation(), directoryName)

	err := influxClient.BackupInflux(backupParams)
	if err != nil {
		return err
	}

	err = tarutils.Tar(backupParams.Path, r.storageDriver.GetLocalStorageLocation())
	if err != nil {
		return err
	}

	err = r.storageDriver.UploadTar(fmt.Sprintf("%s.tar", directoryName))
	if err != nil {
		return err
	}

	return nil
}

func cleanupBackup(backupPath string) error {
	errBackupPath := os.RemoveAll(backupPath)
	errTar := os.RemoveAll(fmt.Sprintf("%s.tar", backupPath))

	if errBackupPath != nil && errTar != nil {
		return errors.Wrapf(errBackupPath, errBackupPath.Error())
	} else if errBackupPath != nil {
		return errBackupPath
	} else if errTar != nil {
		return errTar
	}

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
