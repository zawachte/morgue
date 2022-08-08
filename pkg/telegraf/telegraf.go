package telegraf

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/BurntSushi/toml"
)

type TelegrafConfig struct {
	Token        string
	Urls         []string
	Organization string
	Bucket       string
}

func WriteTelegrafConfig(config TelegrafConfig, path string) error {
	buf := new(bytes.Buffer)
	top := map[string]interface{}{
		"global_tags": map[string]interface{}{},
		"agent": map[string]interface{}{
			"interval":            "10s",
			"round_interval":      true,
			"metric_batch_size":   1000,
			"metric_buffer_limit": 10000,
			"collection_jitter":   "0s",
			"flush_interval":      "10s",
			"flush_jitter":        "0s",
			"precision":           "0s",
			"hostname":            "",
			"omit_hostname":       false,
		},
		"outputs": map[string]interface{}{
			"influxdb_v2": map[string]interface{}{
				"urls":         config.Urls,
				"token":        config.Token,
				"organization": config.Organization,
				"bucket":       config.Bucket,
			},
		},
		"inputs": map[string]interface{}{
			"cpu": map[string]interface{}{
				"percpu":           true,
				"totalcpu":         true,
				"collect_cpu_time": false,
				"report_active":    false,
				"core_tags":        false,
			},
			"disk": map[string]interface{}{
				"ignore_fs": []string{"tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"},
			},
			"diskio":    map[string]interface{}{},
			"kernel":    map[string]interface{}{},
			"processes": map[string]interface{}{},
			"swap":      map[string]interface{}{},
			"system":    map[string]interface{}{},
		},
	}

	if err := toml.NewEncoder(buf).Encode(top); err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, buf.Bytes(), 0777); err != nil {
		return err
	}

	return nil
}

func RunTelegraf(abort <-chan error, config TelegrafConfig, telegrafLocation string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fullPath := path.Join(wd, "telegraf.conf")

	err = WriteTelegrafConfig(config, fullPath)
	if err != nil {
		return err
	}
	defer os.Remove(fullPath)

	configArg := []string{"--config", fullPath}

	/* #nosec */
	cmd := exec.Command(telegrafLocation, configArg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-abort:
		if errKill := cmd.Process.Kill(); errKill != nil {
		}

		return err
	case err := <-done:
		return err
	}
}
