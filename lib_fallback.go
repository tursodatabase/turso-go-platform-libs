//go:build !(windows && amd64) && !(linux && arm64) && !(linux && amd64) && !(darwin && arm64) && !(darwin && amd64) && !(windows && arm64)

package turso_libs

import "embed"

var libs embed.FS
