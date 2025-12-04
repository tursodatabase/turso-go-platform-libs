//go:build windows && amd64

package turso_libs

import "embed"

//go:embed libs/windows_amd64/*
var libs embed.FS
