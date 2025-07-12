package clickhouse

import "github.com/google/uuid"

type ClickhouseDevice struct {
	ID      uuid.UUID `ch:"id"`
	Label   string    `ch:"label"`
	Secret  string    `ch:"secret"`
	Version int32     `ch:"version"`
}
