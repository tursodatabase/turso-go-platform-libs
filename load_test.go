package turso_libs

import (
	"testing"

	"github.com/ebitengine/purego"
	"github.com/stretchr/testify/require"
)

var (
	turso_version func() string
)

func TestLoad(t *testing.T) {
	t.Log(libraryFilename())
	t.Log(embeddedLibraryPath())

	library, err := LoadTursoLibrary(LoadTursoLibraryConfig{LoadStrategy: "mixed"})
	require.Nil(t, err)

	purego.RegisterLibFunc(&turso_version, library, "turso_version")
	t.Log("turso_version", turso_version())
	require.NotEmpty(t, turso_version())
}
