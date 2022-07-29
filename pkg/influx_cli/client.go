package influx_cli

import (
	"context"
	"fmt"
	"net/url"
	"runtime"

	influxapi "github.com/influxdata/influx-cli/v2/api"
	"github.com/influxdata/influx-cli/v2/clients"
	"github.com/influxdata/influx-cli/v2/clients/backup"
	"github.com/influxdata/influx-cli/v2/clients/setup"
	"github.com/influxdata/influx-cli/v2/config"

	"github.com/influxdata/influx-cli/v2/pkg/stdio"
)

// newCli builds a CLI core that reads from stdin, writes to stdout/stderr, manages a local config store,
// and optionally tracks a trace ID specified over the CLI.
func newCli() (clients.CLI, error) {
	configPath, err := config.DefaultPath()
	if err != nil {
		return clients.CLI{}, err
	}

	configSvc := config.NewLocalConfigService(configPath)

	activeConfig, err := configSvc.Active()
	if err != nil {
		return clients.CLI{}, err
	}

	return clients.CLI{
		StdIO:            stdio.TerminalStdio,
		PrintAsJSON:      true,
		HideTableHeaders: false,
		ActiveConfig:     activeConfig,
		ConfigService:    configSvc,
	}, nil
}

// newApiClient returns an API clients configured to communicate with a remote InfluxDB instance over HTTP.
// Client parameters are pulled from the CLI context.
func newApiClient(configSvc config.Service, injectToken bool) (*influxapi.APIClient, error) {
	cfg, err := configSvc.Active()
	if err != nil {
		return nil, err
	}

	configParams := influxapi.ConfigParams{
		UserAgent:        fmt.Sprintf("influx/%s", runtime.GOOS),
		AllowInsecureTLS: false,
		Debug:            true,
	}

	parsedHost, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("host URL %q is invalid: %w", cfg.Host, err)
	}
	configParams.Host = parsedHost

	if injectToken {
		configParams.Token = &cfg.Token
	}

	apiConfig := influxapi.NewAPIConfig(configParams)

	return influxapi.NewAPIClient(apiConfig), nil
}

type Client interface {
	SetupInflux(SetupInfluxParams) error
	BackupInflux(BackupInfluxParams) error
}

type client struct {
	cli       clients.CLI
	apiClient *influxapi.APIClient
}

func NewClient() (Client, error) {
	cli, err := newCli()
	if err != nil {
		return nil, err
	}

	apiClient, err := newApiClient(cli.ConfigService, true)
	if err != nil {
		return nil, err
	}

	return &client{
		cli:       cli,
		apiClient: apiClient,
	}, nil
}

type SetupInfluxParams struct {
	Username  string
	Password  string
	AuthToken string
	Org       string
	Bucket    string
	Retention string
}

func (c *client) SetupInflux(inputParams SetupInfluxParams) error {

	client := setup.Client{
		CLI:      c.cli,
		SetupApi: c.apiClient.SetupApi,
	}

	params := setup.Params{
		Username:  inputParams.Username,
		Password:  inputParams.Password,
		AuthToken: inputParams.AuthToken,
		Org:       inputParams.Org,
		Bucket:    inputParams.Bucket,
		Retention: inputParams.Retention,
		Force:     true,
	}

	err := client.Setup(context.Background(), &params)
	if err != nil {
		return err
	}

	return nil
}

type BackupInfluxParams struct {
	Org    string
	Bucket string
	Path   string
}

func (c *client) BackupInflux(inputParams BackupInfluxParams) error {

	client := backup.Client{
		CLI:       c.cli,
		BackupApi: c.apiClient.BackupApi,
		HealthApi: c.apiClient.HealthApi,
	}

	params := backup.Params{
		Path:        inputParams.Path,
		Compression: 1,
	}

	params.BucketName = inputParams.Bucket
	params.OrgName = inputParams.Org

	err := client.Backup(context.Background(), &params)
	if err != nil {
		return err
	}

	return nil
}
