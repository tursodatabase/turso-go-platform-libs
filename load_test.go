package turso_libs

import (
	"strings"
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
	library, err := LoadTursoLibrary(LoadTursoLibraryConfig{LoadStrategy: "system"})
	if err != nil && strings.Contains(err.Error(), "not found") {
		t.Skipf("no library found")
	}
	require.Nil(t, err)

	purego.RegisterLibFunc(&turso_version, library, "turso_version")
	t.Log("turso_version", turso_version())
	require.NotEmpty(t, turso_version())
}
