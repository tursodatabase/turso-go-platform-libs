//go:build darwin && amd64

package turso_libs

import "embed"

//go:embed libs/darwin_amd64/*
var libs embed.FS
