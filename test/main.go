package main

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

func main() {
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
		"outputs.influxdb_v2": map[string]interface{}{
			"urls":         []string{"dd"},
			"token":        "dd",
			"organization": "dd",
			"bucket":       "ddd",
		},
		"inputs.cpu": map[string]interface{}{
			"percpu":           true,
			"totalcpu":         true,
			"collect_cpu_time": false,
			"report_active":    false,
			"core_tags":        false,
		},
		"inputs.disk": map[string]interface{}{
			"ignore_fs": []string{"tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"},
		},
		"inputs.diskio":    map[string]interface{}{},
		"inputs.kernel":    map[string]interface{}{},
		"inputs.processes": map[string]interface{}{},
		"inputs.swap":      map[string]interface{}{},
		"inputs.system":    map[string]interface{}{},
	}

	if err := toml.NewEncoder(buf).Encode(top); err != nil {
		panic(err)
	}

	fmt.Print(string(buf.Bytes()))
}
