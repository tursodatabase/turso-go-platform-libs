//go:build windows && arm64

package turso_libs

import "embed"

//go:embed libs/windows_arm64/*
var libs embed.FS
