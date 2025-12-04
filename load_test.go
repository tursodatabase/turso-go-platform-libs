package turso_libs

import "testing"

func TestLoad(t *testing.T) {
	t.Log(LoadTursoLibrary(LoadTursoLibraryConfig{LoadStrategy: ""}))
}
