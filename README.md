# redfish_exporter
A prometheus expoter to get  metrics from redfish based servers such as lenovo/dell/Supermicro servers.

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
```
Note that the ```default`` entry is useful as it avoids an error
condition that is discussed in [this issue][2].

## Building

To build the redfish_exporter executable run the command:
```sh
make build
```

or build in centos 7 docker image
```sh
make docker-build-centos7
```

or build in centos 8 docker image
```sh
make docker-build-centos8
```

## Running

To run redfish_exporter do something like:
```sh
redfish_exporter --config.file=redfish_exporter.yml
```
and run
```sh
redfish_exporter -h
```
for more options.

## Scraping

We can get the metrics via
```
curl http://<redfish_exporter host>:9610/redfish?target=10.36.48.24

```
or by pointing your favourite browser at this URL.

## prometheus job conf

You can then setup [Prometheus][3] to scrape the target using
something like this in your Prometheus configuration files:
```yaml
  - job_name: 'redfish-exporter'

    # metrics_path defaults to '/metrics'
    metrics_path: /redfish

    # scheme defaults to 'http'.

    static_configs:
    - targets:
       - 10.36.48.24
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9610  ### the address of the redfish-exporter address
```
Note that port 9610 has been [reserved][4] for the redfish_exporter.
## Supported Devices (tested)
- Lenovo ThinkSystem SR850 (BMC 2.1/2.42)
- Lenovo ThinkSystem SR650 (BMC 2.50)

## Acknowledgement

- [gofish][5] provides the underlying library to interact servers

[1]: git@github.com:sbates130272/redfish_exporter.git
[2]: https://github.com/jenningsloy318/redfish_exporter/issues/7
[3]: https://prometheus.io/
[4]: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
[5]: https://github.com/stmcginnis/gofish
