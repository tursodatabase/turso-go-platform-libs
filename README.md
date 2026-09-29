# turso-go-platform-libs
pre-compiled platform-specific libraries for tursodb go driver

## FreeBSD

FreeBSD has no prebuilt library in this package yet. To use it:

1. **Install Rust and Cargo** on FreeBSD (run the package command as root):

   ```sh
   su -c 'pkg install rust'
   cargo --version
   ```

2. **Build the shared library on FreeBSD** from this repository. Set
   `TURSO_RS_BUILD_REF` to the Turso tag or branch you want to build:

   ```sh
   TURSO_RS_BUILD_REF=main ./build_libs.sh
   ```

   The result is `libs/freebsd_amd64/libturso_sync_sdk_kit.so` on amd64, or
   `libs/freebsd_arm64/libturso_sync_sdk_kit.so` on arm64.

3. **Copy the library to `/usr/local/lib/turso`** from this repository. On amd64:

   ```sh
   su -c 'mkdir -p /usr/local/lib/turso && cp libs/freebsd_amd64/libturso_sync_sdk_kit.so /usr/local/lib/turso/'
   ```

   On arm64, replace `freebsd_amd64` with `freebsd_arm64`. Set the search path
   in the shell that runs your Go program:

   ```sh
   export LD_LIBRARY_PATH=/usr/local/lib/turso
   ```

4. **Point your app at this checkout** of the FreeBSD-support branch. From your app's Go module directory, if this repository is in your home folder, run:

   ```sh
   go mod edit -replace github.com/tursodatabase/turso-go-platform-libs="$HOME/turso-go-platform-libs"
   ```

   Adjust the path if you cloned it elsewhere. This local `replace` is needed because `tursogo` otherwise pulls the published platform-libs module, which does not contain the FreeBSD loader changes.

With this checkout, FreeBSD defaults to loading the system library. Your app can
import `turso.tech/database/tursogo` normally; no FreeBSD-specific Go code or
`InitLibrary` call is needed. Other platforms retain their embedded-library default.

If you **cross-compile with `CGO_ENABLED=0`** for FreeBSD, purego v0.9.1 also requires this Go build flag:

```sh
GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -gcflags=github.com/ebitengine/purego/internal/fakecgo=-std ./...
```

Use `GOARCH=arm64` instead for FreeBSD arm64.
