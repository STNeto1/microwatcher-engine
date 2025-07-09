package clickhouse

type ClickhouseDevice struct {
	ID     string `ch:"id"`
	Label  string `ch:"label"`
	Secret string `ch:"secret"`
}
