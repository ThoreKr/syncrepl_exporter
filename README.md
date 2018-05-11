# syncrepl_exporter

[Prometheus](https://prometheus.io/) exporter for OpenLDAP sync replication stats.

## Getting

```
$ go get github.com/ThoreKr/syncrepl_exporter
```

## Building

```
$ cd $GOPATH/github.com/ThoreKr/syncrepl_exporter
$ make
```

## Configuration

See [config.example.yam](config.example.yaml), an illustrative example of the configuration file in YAML format.


## Running

```
$ ./syncrepl_exporter <flags>
```

```
$ ./syncrepl_exporter -h
usage: syncrepl_exporter [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9328"
                          Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                          Path under which to expose metrics.
      --path.config="config.yaml"
                          Configuration YAML file path.
      --log.level="info"  Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"
                          Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version           Show application version.
```

