module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/influxdbexporter

go 1.16

require (
	github.com/influxdata/influxdb-observability/common v0.1.0
	github.com/influxdata/influxdb-observability/otel2influx v0.1.0
	github.com/influxdata/line-protocol/v2 v2.0.0-20210428091617-0567a5134992
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210701204331-d1fced9688ba
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.19.1
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210701204331-d1fced9688ba
