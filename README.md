# redfish_exporter
expoter to get  metrics from redfish based servers such as lenovo servers



example configure set as [example](./scripts/redfish_exporter.yml)
```yaml
hosts:
    10.36.48.24:
      username: admin
      password: pass
```



then start redfish_exporter via 
```sh
redfish_exporter --config.file=redfish_exporter.yml
```

then we can get the metrics via 
```
curl http://<redfish_exporter host>:9610/redfish?target=10.36.48.24

```

## prometheus job conf
add redfish-exporter job conif as following
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
````


## Supported Devices(tested)
- Lenovo ThinkSystem SR850 (BMC 2.1/2.42)
- Lenovo ThinkSystem SR650 (BMC 2.50)

## Acknowledgement

- [gofish](github.com/stmcginnis/gofish) provides the underlying library to interact servers  