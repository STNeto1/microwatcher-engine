package cli

type Start struct {
	MetricInterval      string `help:"Interval between runs" default:"5s"`
	HealthCheckInterval string `help:"Interval between health checks" default:"5s"`
	Identifier          string `help:"Identifier used to identify this device, defaults to hostname" default:""`
	ClientID            string `help:"Client ID used to sign telemetry" default:""`
	ClientSecret        string `help:"Secret used to sign telemetry" default:""`
}

type Check struct {
	ClientID     string `help:"Client ID used to sign telemetry" default:""`
	ClientSecret string `help:"Secret used to sign telemetry" default:""`
}

type CLI struct {
	Start Start `cmd:"" help:"Start agent"`
	Check Check `cmd:"" help:"Check agent connection"`
}
