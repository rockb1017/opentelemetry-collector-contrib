module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/alibabacloudlogserviceexporter

go 1.16

require (
	github.com/aliyun/aliyun-log-go-sdk v0.1.21
	github.com/census-instrumentation/opencensus-proto v0.3.0
	github.com/gogo/protobuf v1.3.2
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210701204331-d1fced9688ba
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.19.1
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210701204331-d1fced9688ba
