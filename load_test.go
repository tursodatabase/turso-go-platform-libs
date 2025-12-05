package turso_libs

import (
	"path"
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

	var list func(p string)
	list = func(p string) {
		file, err := libs.Open(p)
		t.Log("open", p, err)
		if err != nil {
			return
		}

		stat, err := file.Stat()
		t.Log("stat", stat, err)
		if err != nil {
			return
		}

		if !stat.IsDir() {
			return
		}

		dirs, err := libs.ReadDir(p)
		t.Log("read_dir", dirs, err)
		if err != nil {
			return
		}

		for _, dir := range dirs {
			list(path.Join(p, dir.Name()))
		}
	}
	list(".")

	library, err := LoadTursoLibrary(LoadTursoLibraryConfig{LoadStrategy: "mixed"})
	require.Nil(t, err)

	purego.RegisterLibFunc(&turso_version, library, "turso_version")
	t.Log("turso_version", turso_version())
	require.NotEmpty(t, turso_version())
}
