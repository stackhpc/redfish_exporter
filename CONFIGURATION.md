# Configuration

Configuration is provided by a yml file with the path passed like `--config.file config.yml` or by mounting the config to `/etc/prometheus/redfish_exporter.yml` in a Docker container running the provided image.

Configuration can be provided either per host or per group, and default configs can be provided with the `default` host.
It's important to note that configs are *not* merged. Configs are found in this order and the process stops when one is found:

1. Group config
2. Host config
3. Default host config

This means that you cannot provide a partial configuration for a specific host and expect the other configuration values to be filled from the default config.

## Example configuration

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

Note that the `default` entry is useful as it avoids an error
condition that is discussed in [this issue][1].

## Log collection

One issue you may encounter is slow scrape times when collecting logs from
BMCs. Sometimes these scrapes can take ~10 minutes. If this becomes
problematic the collection of logs can be configured as follows:

1. via the the config file

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

2. via the `collectlogs` query parameter

For example:

```sh
curl '127.0.0.1:9123/redfish?target=10.10.12.23&collectlogs=false'
```

3. Via a combination of both options

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

```yaml
hosts:
  default:
    username: username
    password: password
    collectlogs: true
    logcount:
      DEFAULT: 10
```

For a particular log service, `Sel`:

```yaml
hosts:
  default:
    username: username
    password: password
    collectlogs: true
    logcount:
      Sel: 20
```

Collect all logs for one service, `Sel` , but limit others:

```yaml
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

## Configuring current requests

By default, up to 128 HTTP requests can be in-flight to a BMC concurrently.
Some BMCs may be able to handle more, in which case you can increase this value to get faster scrape times.
Some BMCs may be able to handle less, which will show up in the exporter logs as i/o timeouts. In this case,
the value should be reduced at the cost of longer scrape times.

This can be done by setting `concurrent_requests` for a group or host.

```yaml
hosts:
  default:
    concurrent_requests: 32
```

## Configuring auth backoff

Some BMCs issue IP bans when clients fail to auth. This can cause issues if other services on the same IP need to access the BMC.
By default, the exporter will not reconnect to a BMC if it failed to connect within the last 60 mins.
This value can be overridden by setting `auth_backoff_secs` for a group or host.

```yaml
hosts:
  default:
    auth_backoff_secs: 7200
```

## Reloading Configuration

```
PUT /-/reload
POST /-/reload
```

The `/-/reload` endpoint triggers a reload of the redfish_exporter configuration.
500 will be returned when the reload fails.

Alternatively, a configuration reload can be triggered by sending `SIGHUP` to the redfish_exporter process as well.

## Prometheus Configuration

You can then setup Prometheus to scrape the target using
something like this in your Prometheus configuration files:

```yaml
- job_name: "redfish-exporter"

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
      replacement: localhost:9610 ### the address of the redfish-exporter address, hence relpace localhost with the server IP address that redfish-export is running on
    # (optional) when using group config add this to have group=my_group_name
    - target_label: __param_group
      replacement: my_group_name
```

[1]: https://github.com/jenningsloy318/redfish_exporter/issues/7
