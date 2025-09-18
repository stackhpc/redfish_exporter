# redfish_exporter
A prometheus exporter to get metrics from redfish based servers.

This is a fork of the original project from https://github.com/jenningsloy318/redfish_exporter

## Configuration

An example configure given as an [example][1]:
```yaml
hosts:
  10.36.48.24:
    username: admin
    password: pass
  default:
    username: admin
    password: pass
groups:
  group1:
    username: group1_user
    password: group1_pass
  flakey:
    username: "Administrator"
    password: "Password"
    disabled_metrics:
      - "ethernet_interfaces"
```
Note that the ```default``` entry is useful as it avoids an error
condition that is discussed in [this issue][2].

## Log collection

One issue you may encounter is slow scrape times when collecting logs from
BMCs. Sometimes these scrapes can take ~10minutes. If this becomes
problematic the collection of logs can be configured as follows:

1) via the the config file

To disable log collection you can set:

```yaml
hosts:
  default:
    username: username
    password: password
    collectlogs: false
```

or for a group:

```yaml
groups:
  group1:
    username: username
    password: password
    collectlogs: false
```

2) via the `collectlogs` query parameter

For example:

```sh
curl '127.0.0.1:9123/redfish?target=10.10.12.23&collectlogs=false'
```

3) Via a combination of both options

For example, collect metrics without logs by default:

Set:

```yaml
collectlogs: false
```

Every hour or so, override the config setting and fetch the logs:

```sh
curl '127.0.0.1:9123/redfish?target=10.10.12.23&collectlogs=true'
```

The `collectlogs` query parameter can be included in Prometheus config.

### Log collection count

To restrict the number of logs collected for all log services:

```
hosts:
  default:
    username: username
    password: password
    collectlogs: true
    logcount:
      DEFAULT: 10
```

For a particular log service, `Sel`:

```
hosts:
  default:
    username: username
    password: password
    collectlogs: true
    logcount:
      Sel: 20
```

Collect all logs for one service, `Sel` , but limit others:

```
hosts:
  default:
    username: username
    password: password
    collectlogs: true
    logcount:
      DEFAULT: 10
      Sel: -1
```

## Disabling collection of specific groups of metrics

Sometimes it isn't possible to gather all metrics from a BMC because
it doesn't conform to the Redfish standard due to a bug. Or perhaps
you don't need all the metrics. In either case, basic support exists
for skipping specific metric groups at the `default`, `group` or
`host` level:

```yaml
hosts:
  10.36.48.24:
    username: admin
    password: pass
    disabled_metrics:
      - "ethernet_interfaces"
  default:
    username: admin
    password: pass
    disabled_metrics:
      - "memory"
      - "processor"
      - "storage"
      - "pcie_devices"
      - "network_interfaces"
      - "ethernet_interfaces"
      - "simple_storage"
      - "pcie_functions"
```

## Building

To build the redfish_exporter executable run the command:
```sh
make build
```

## Running
- running directly on linux
  ```sh
  redfish_exporter --config.file=redfish_exporter.yml
  ```
  and run   `redfish_exporter -h
  `  for more options.

- running in container
  
  Also if you build it as a docker image, you can also run in container, just remember to replace your config  `/etc/prometheus/redfish_exporter.yml` in container
## Scraping

We can get the metrics via
```
curl http://<redfish_exporter host>:9610/redfish?target=10.36.48.24

```
or by pointing your favourite browser at this URL.

## Reloading Configuration
```
PUT /-/reload
POST /-/reload
```
The `/-/reload` endpoint triggers a reload of the redfish_exporter configuration.
500 will be returned when the reload fails.

Alternatively, a configuration reload can be triggered by sending `SIGHUP` to the redfish_exporter process as well.

## Prometheus Configuration

You can then setup [Prometheus][3] to scrape the target using
something like this in your Prometheus configuration files:
```yaml
  - job_name: 'redfish-exporter'

    # metrics_path defaults to '/metrics'
    metrics_path: /redfish

    # scheme defaults to 'http'.

    static_configs:
    - targets:
       - 10.36.48.24 ## here is the list of the redfish targets which will be monitored
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9610  ### the address of the redfish-exporter address, hence relpace localhost with the server IP address that redfish-export is running on
      # (optional) when using group config add this to have group=my_group_name
      - target_label: __param_group
        replacement: my_group_name
```
Note that port 9610 has been [reserved][4] for the redfish_exporter.
## Supported Devices (tested)
- Enginetech EG520R-G20 (Supermicro Firmware Revision 1.76.39)
- Enginetech EG920A-G20 (Huawei iBMC 6.22)
- Lenovo ThinkSystem SR850 (BMC 2.1/2.42)
- Lenovo ThinkSystem SR650 (BMC 2.50)
- Dell PowerEdge R440, R640, R650, R6515, C6420
- GIGABYTE G292-Z20, G292-Z40, G482-Z54

## Acknowledgement

- [gofish][5] provides the underlying library to interact servers

[1]: git@github.com:sbates130272/redfish_exporter.git
[2]: https://github.com/jenningsloy318/redfish_exporter/issues/7
[3]: https://prometheus.io/
[4]: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
[5]: https://github.com/stmcginnis/gofish
