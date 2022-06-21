# OTLP HTTP Experiment

This code creates some mock metrics and exports them to `stdout` or an HTTP endpoint.

## Run
The default exporter is `stdout`. To change to the HTTP exporter, set the `METRICS_EXPORTER` environment variable:
```
export METRICS_EXPORTER=http
```
and update the global `collectorEndpoint` variable.  
To run:
```
go run main.go
```
