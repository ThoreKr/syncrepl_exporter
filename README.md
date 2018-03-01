# syncrepl_exporter
Prometheus exporter for syncrepl stats

```
Usage of syncrepl_exporter:
  -base.dn string
    	'dc=example,dc=org' the base DN of the directory
  -ldap.host string
    	hostname:port of the ldap server (default "localhost:636")
  -telemetry.addr string
    	host:port for syncrepl exporter (default ":9328")
  -telemetry.path string
    	URL path for surfacing collected metrics (default "/metrics")
```
