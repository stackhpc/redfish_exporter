# redfish_exporter
A Prometheus exporter to get metrics from Redfish-based servers.

This is a largely reworked fork of the original project from [jenningsloy318](https://github.com/jenningsloy318/redfish_exporter).
It supports more metrics, including the new `ThermalSubsystem` and `PowerSubsystem` resources.
Scrape performance is greatly improved, and unit testing is employed to ensure stability.
Metrics present in the original exporter should still be present here, see [v3.0.0 release notes][2] for more details on potentially breaking changes.

## Building

To build the `redfish_exporter` executable run the command:
```sh
make build
```

## Configuration
See [CONFIGURATION.md](CONFIGURATION.md)

## Running
### As a binary
```sh
redfish_exporter --config.file=redfish_exporter.yml
```
and run `redfish_exporter -h` for more options.

### As a container
An official container image is provided at `ghcr.io/stackhpc/redfish_exporter`.

You should mount your config file to `/etc/prometheus/redfish_exporter.yml` and expose port `9610`.

## Scraping

You can get the metrics via
```
curl http://<redfish_exporter host>:9610/redfish?target=10.36.48.24
```
or by pointing your favourite browser at this URL.

## Supported Devices (tested)

Since this exporter complies with the Redfish standard it should work on any Redfish-compliant device.
If you encounter any issues please don't hesitate to open an issue or a PR.

The below servers have been tested with the exporter.

- Enginetech EG520R-G20 (Supermicro Firmware Revision 1.76.39)
- Enginetech EG920A-G20 (Huawei iBMC 6.22)
- Lenovo ThinkSystem SR850 (BMC 2.1/2.42)
- Lenovo ThinkSystem SR650 (BMC 2.50)
- Dell PowerEdge R440, R640, R650, R6515, C6420
- GIGABYTE G292-Z20, G292-Z40, G482-Z54

## Acknowledgement

- [gofish][1] provides the underlying library to interact servers

[1]: https://github.com/stmcginnis/gofish
[2]: https://github.com/stackhpc/redfish_exporter/releases/tag/v3.0.0
