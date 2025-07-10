module github.com/microwatcher/agent

go 1.24.4

require (
	github.com/alecthomas/kong v1.12.0
	github.com/microwatcher/shared v0.0.0-00010101000000-000000000000
	github.com/shirou/gopsutil v3.21.11+incompatible
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

replace github.com/microwatcher/shared => ../shared

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
)
