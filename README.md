# morgue

morgue configures telegraf and influxd to collect system metrics for a linux machine and take and upload backups of influxdb and optionally upload them to the cloud.

This is useful in IOT and edge cases when machines don't have relible internet connectivity to emit metrics on 10-20s intervals. 

It is also useful if you have a cloud instance with monitoring tied to kubernetes, and the node goes down and some metrics data is lost.

## Getting Started

### Prerequisites

#### Operating Systems
morgue only runs on linux. 

#### Backup Storage

The only storage locations for backups are local storage and s3. For s3 storage you need to ensure that the targeted bucket already exists.

### Systemd service mode

Install the rpms for influxdb and telegraf.

influxdb2:
```
# Ubuntu/Debian
wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.3.0-amd64.deb
sudo dpkg -i influxdb2-2.3.0-amd64.deb

# Red Hat/CentOS/Fedora
wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.3.0-amd64.rpm
sudo yum localinstall influxdb2-2.3.0-amd64.rpm
```

telegraf:
```
# Ubuntu/Debian
wget -q https://repos.influxdata.com/influxdb.key
echo '23a1c8836f0afc5ed24e0486339d7cc8f6790b83886c4c96995b88a061c5bb5d influxdb.key' | sha256sum -c && cat influxdb.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdb.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdb.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list
sudo apt-get update && sudo apt-get install telegraf


# Red Hat/CentOS/Fedora
cat <<EOF | sudo tee /etc/yum.repos.d/influxdata.repo
[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/\$basearch/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdb.key
EOF

sudo yum install telegraf
```

Now build morgue:
```
make
```

Then copy the morgue binary to the `/usr/bin/morgue` on the target system. After, you will need to edit the unit file at `config/systemd/morgue.service` to your desired configuration.  After editing, you will need to copy the unit file to `/etc/systemd/system/morgue.service`.

Lastly, you will need to write a file at `/etc/default/morgue` with something like the following:

```
OPTIONS="--service-mode=true --backup-frequency 1h --storage-driver aws --aws-region us-east-1 --aws-s3-bucket samples-metrics-bucket"
```

Now we are ready to start the service:

```
systemctl start morgue
```

### Embedded mode

Embedded mode is not suggested for production use but very useful for quickly deploying morgue.

First install the influxdb binary:

```
make influxd
```

Then install the telegraf binary:
```
make telegraf
```

Then build morgue
```
make morgue
```

And run the binary:

```sh
./bin/morgue --backup-frequency 1h \
 --storage-driver aws \
 --aws-region us-east-1 \
 --aws-s3-bucket samples-metrics-bucket
```

## Consuming the backups

COMING soon: `morguectl`: tooling to simpify extraction and loading of the influxdb backups to a fresh influxdb and local grafana UI.