GOCMD=go
GOBUILD=$(GOCMD) build
GOHOSTOS=$(strip $(shell $(GOCMD) env get GOHOSTOS))

TAG ?= $(shell git describe --tags)
COMMIT ?= $(shell git describe --always)
BUILD_DATE ?= $(shell date -u +%m/%d/%Y)


MORGUE=bin/morgue

all: target

clean:
	rm -rf ${MORGUE} 

target:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -ldflags "-X main.version=$(TAG) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)" -o ${MORGUE} github.com/zawachte/morgue

influxd:
	wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.2.0-linux-amd64.tar.gz
	tar xvzf influxdb2-2.2.0-linux-amd64.tar.gz
	cp influxdb2-2.2.0-linux-amd64/influxd bin/
	rm -r influxdb2-2.2.0-linux-amd64
	rm influxdb2-2.2.0-linux-amd64.tar.gz

telegraf:
	wget https://dl.influxdata.com/telegraf/releases/telegraf-1.23.3_linux_amd64.tar.gz
	tar xvzf telegraf-1.23.3_linux_amd64.tar.gz
	cp ./telegraf-1.23.3/usr/bin/telegraf bin/
	rm -r telegraf-1.23.3
	rm telegraf-1.23.3_linux_amd64.tar.gz