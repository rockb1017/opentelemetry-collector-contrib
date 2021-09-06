module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/influxdata/influxdb-observability/common v0.2.6
	github.com/influxdata/influxdb-observability/influx2otel v0.2.6
	github.com/influxdata/line-protocol/v2 v2.0.0-20210428091617-0567a5134992
	go.opentelemetry.io/collector v0.29.1-0.20210701204331-d1fced9688ba
	go.opentelemetry.io/collector/model v0.34.0
	go.uber.org/zap v1.18.1
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210701204331-d1fced9688ba
