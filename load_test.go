package turso_libs

import (
	"path"
	"path/filepath"
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

func TestEmbeddedLibraryFTS(t *testing.T) {
	library, err := LoadTursoLibrary(LoadTursoLibraryConfig{LoadStrategy: EmbeddedLibraryLoadStrategy})
	require.NoError(t, err)

	type databaseConfig struct {
		asyncIO              uint64
		path                 *byte
		experimentalFeatures *byte
		vfs                  *byte
		encryptionCipher     *byte
		encryptionHexkey     *byte
		pageCodec            uintptr
		openFlags            uint32
	}
	var databaseNew func(*databaseConfig, *uintptr, **byte) int32
	var databaseOpen func(uintptr, **byte) int32
	var databaseConnect func(uintptr, *uintptr, **byte) int32
	var databaseDeinit func(uintptr)
	var connectionPrepare func(uintptr, string, *uintptr, **byte) int32
	var connectionDeinit func(uintptr)
	var statementStep func(uintptr, **byte) int32
	var statementRunIO func(uintptr, **byte) int32
	var statementValueInt func(uintptr, uintptr) int64
	var statementDeinit func(uintptr)
	purego.RegisterLibFunc(&databaseNew, library, "turso_database_new")
	purego.RegisterLibFunc(&databaseOpen, library, "turso_database_open")
	purego.RegisterLibFunc(&databaseConnect, library, "turso_database_connect")
	purego.RegisterLibFunc(&databaseDeinit, library, "turso_database_deinit")
	purego.RegisterLibFunc(&connectionPrepare, library, "turso_connection_prepare_single")
	purego.RegisterLibFunc(&connectionDeinit, library, "turso_connection_deinit")
	purego.RegisterLibFunc(&statementStep, library, "turso_statement_step")
	purego.RegisterLibFunc(&statementRunIO, library, "turso_statement_run_io")
	purego.RegisterLibFunc(&statementValueInt, library, "turso_statement_row_value_int")
	purego.RegisterLibFunc(&statementDeinit, library, "turso_statement_deinit")

	path := append([]byte(filepath.Join(t.TempDir(), "fts.db")), 0)
	features := append([]byte("index_method"), 0)
	config := databaseConfig{path: &path[0], experimentalFeatures: &features[0]}
	var database, connection uintptr
	require.Zero(t, databaseNew(&config, &database, nil))
	defer databaseDeinit(database)
	require.Zero(t, databaseOpen(database, nil))
	require.Zero(t, databaseConnect(database, &connection, nil))
	defer connectionDeinit(connection)

	for _, sql := range []string{
		"CREATE TABLE docs (id INTEGER PRIMARY KEY, body TEXT)",
		"INSERT INTO docs VALUES (1, 'searchable document'), (2, 'another entry')",
		"CREATE INDEX docs_fts ON docs USING fts (body)",
		"SELECT id FROM docs WHERE fts_match(body, 'searchable')",
	} {
		var statement uintptr
		require.Zero(t, connectionPrepare(connection, sql, &statement, nil), sql)
		func() {
			defer statementDeinit(statement)
			rows := 0
			for {
				switch status := statementStep(statement, nil); status {
				case 1: // done
					if strings.HasPrefix(sql, "SELECT") {
						require.Equal(t, 1, rows)
					}
					return
				case 2: // row
					rows++
					require.EqualValues(t, 1, statementValueInt(statement, 0))
				case 3: // IO
					require.Zero(t, statementRunIO(statement, nil), sql)
				default:
					t.Fatalf("%s: unexpected statement status %d", sql, status)
				}
			}
		}()
	}
}
