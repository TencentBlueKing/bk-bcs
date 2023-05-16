package db_migration

import "embed"

//go:embed migrations/sql
var SQLFiles embed.FS
