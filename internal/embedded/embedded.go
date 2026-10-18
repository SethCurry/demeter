package embedded

import (
	"embed"
)

//go:embed schema.sql
var SQLSchema string

//go:embed ui/*
var UI embed.FS
