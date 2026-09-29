# turso-go-platform-libs
pre-compiled platform-specific libraries for tursodb go driver

## FreeBSD

FreeBSD has no prebuilt library in this package yet. To use it:

1. **Build the shared library on FreeBSD** from this repository. Set
   `TURSO_RS_BUILD_REF` to the Turso tag or branch you want to build:

   ```sh
   TURSO_RS_BUILD_REF=<turso-ref> ./build_libs.sh
   ```

   The result is `libs/freebsd_amd64/libturso_sync_sdk_kit.so` on amd64, or
   `libs/freebsd_arm64/libturso_sync_sdk_kit.so` on arm64.

2. **Make the library discoverable** when running your Go program. For example,
   from this repository on amd64:

   ```sh
   export LD_LIBRARY_PATH="$PWD/libs/freebsd_amd64"
   ```

   Alternatively, put the `.so` in your program's working directory.

3. **Select the system library** in your Go program:

   ```go
   import turso_libs "github.com/tursodatabase/turso-go-platform-libs"

   handle, err := turso_libs.LoadTursoLibrary(turso_libs.LoadTursoLibraryConfig{
       LoadStrategy: turso_libs.SystemLibraryLoadStrategy,
   })
   ```

   Handle `err` before using `handle`. The default embedded-only strategy cannot
   work on FreeBSD until an embedded library is distributed.

If you **cross-compile with `CGO_ENABLED=0`** for FreeBSD, purego v0.9.1 also
requires this Go build flag:

```sh
GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -gcflags=github.com/ebitengine/purego/internal/fakecgo=-std ./...
```

Use `GOARCH=arm64` instead for FreeBSD arm64.
