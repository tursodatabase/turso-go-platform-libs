//go:build linux || darwin

package turso_libs

import "github.com/ebitengine/purego"

func loadLibrary(path string) (uintptr, error) {
	// RTLD_NOW - Relocations are performed when the object is loaded.
	// RTLD_GLOBAL - All symbols are available for relocation processing of other modules.
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
