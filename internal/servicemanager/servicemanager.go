package servicemanager

import (
	"os/exec"

	"github.com/zawachte/morgue/pkg/influx"
	"github.com/zawachte/morgue/pkg/influxd"
	"github.com/zawachte/morgue/pkg/telegraf"
	"go.uber.org/zap"
)

type ServiceManager interface {
	RunInfluxD() error
	RunTelegraf(string) error
}

type ServiceManagerParams struct {
	InfluxDLocation  string
	TelegrafLocation string
	Logger           zap.Logger
}

func NewServiceManager(serviceMode bool, params ServiceManagerParams) ServiceManager {
	if serviceMode {
		return &systemDServiceManager{params.Logger}
	}
	return &embeddedServiceManager{
		influxDLocation:  params.InfluxDLocation,
		telegrafLocation: params.TelegrafLocation,
		logger:           params.Logger,
	}
}

type embeddedServiceManager struct {
	influxDLocation  string
	telegrafLocation string
	logger           zap.Logger
}

func (esm *embeddedServiceManager) RunInfluxD() error {
	abortCh := make(chan error, 1)
	go func() {
		err := influxd.RunInfluxD(abortCh, esm.influxDLocation)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (esm *embeddedServiceManager) RunTelegraf(token string) error {
	abortCh := make(chan error, 1)
	go func() {
		err := telegraf.RunTelegraf(abortCh, telegraf.TelegrafConfig{
			Token:        token,
			Urls:         []string{"http://127.0.0.1:8086"},
			Organization: influx.DefaultOrgName,
			Bucket:       influx.DefaultBucketName,
		}, esm.telegrafLocation)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

type systemDServiceManager struct {
	logger zap.Logger
}

func (esm *systemDServiceManager) RunInfluxD() error {
	cmd := exec.Command("systemctl", "reload-or-restart", "influxd")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func (esm *systemDServiceManager) RunTelegraf(token string) error {

	err := telegraf.WriteTelegrafConfig(telegraf.TelegrafConfig{
		Token:        token,
		Urls:         []string{"http://127.0.0.1:8086"},
		Organization: influx.DefaultOrgName,
		Bucket:       influx.DefaultBucketName,
	}, "/etc/telegraf/telegraf.conf")
	if err != nil {
		return nil
	}

	cmd := exec.Command("systemctl", "reload-or-restart", "telegraf")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	return nil
}
