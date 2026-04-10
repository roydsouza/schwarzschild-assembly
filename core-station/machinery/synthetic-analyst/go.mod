module github.com/rds/aethereum-spine/factories/synthetic-analyst

go 1.26.2

replace github.com/rds/aethereum-spine/aethereum-spine => ../../aethereum-spine

require (
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/stretchr/testify v1.11.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
)
