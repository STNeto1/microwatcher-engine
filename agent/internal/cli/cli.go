package cli

type Start struct {
	MetricInterval      string `help:"Interval between runs" default:"5s"`
	Identifier          string `help:"Agent identifier" default:"unknown"`
	HealthCheckInterval string `help:"Interval between health checks" default:"5s"`
}

type CLI struct {
	Start Start    `cmd:"" help:"Start agent"`
	Check struct{} `cmd:"" help:"Check agent connection"`
}
