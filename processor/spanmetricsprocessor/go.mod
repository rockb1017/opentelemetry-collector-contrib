module github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor

go 1.16

require (
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210701204331-d1fced9688ba
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.41.0
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210701204331-d1fced9688ba
