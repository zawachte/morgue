# morgue

morgue configures telegraf and influxd to collect system metrics for a linux machine and take and upload backups of influxdb and optionally upload them to the cloud.

This is useful in IOT and edge cases when machines don't have relible internet connectivity to emit metrics on 10-20s intervals. 

It is also useful if you have a cloud instance with monitoring tied to kubernetes, and the node goes down and some metrics data is lost.

## Getting Started

