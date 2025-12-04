//go:build linux && amd64 && musl

package turso_libs

import "embed"

//go:embed libs/linux_amd64_musl/*
var libs embed.FS
