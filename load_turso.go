package turso_libs

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const LibraryName = "libturso_sync_sdk_kit"

type LibraryLoadStrategy string

const (
	EmbeddedLibraryLoadStrategy LibraryLoadStrategy = "embedded"
	SystemLibraryLoadStrategy   LibraryLoadStrategy = "system"
	MixedLibraryLoadStrategy    LibraryLoadStrategy = "mixed"
)

type LoadTursoLibraryConfig struct {
	// LoadStrategy available options
	// "embedded" or "" (default): always load library from embedded resources
	// "system": always load library from system pathes
	// "mixed": try to load library from the embedded resources and fallback to system search if no library was found
	LoadStrategy LibraryLoadStrategy
}

// loads the library either from embedded artifacts or searching through LD_LIBRARY_PATH list
func LoadTursoLibrary(config LoadTursoLibraryConfig) (handle uintptr, err error) {
	var path string
	switch config.LoadStrategy {
	case EmbeddedLibraryLoadStrategy, "":
		path, err = embeddedLibraryTryCreate()
		if err != nil {
			return 0, fmt.Errorf("failed to load turso library: %w", err)
		}
		if path == "" {
			return 0, fmt.Errorf("turso library is not embedded for the platform in the package")
		}
	case SystemLibraryLoadStrategy:
		path, err = librarySystemSearch()
		if err != nil {
			return 0, err
		}
	case MixedLibraryLoadStrategy:
		path, err = embeddedLibraryTryCreate()
		if err != nil {
			return 0, fmt.Errorf("failed to load turso library: %w", err)
		}
		if path == "" {
			path, err = librarySystemSearch()
			if err != nil {
				return 0, err
			}
		}
	}
	library, err := loadLibrary(path)
	if err != nil {
		return 0, fmt.Errorf("failed to load library from path %v: %w", path, err)
	}
	return library, nil
}

// isMuslLibc detects if the system is using musl libc (Alpine Linux, Void Linux, etc.)
func isMuslLibc() bool {
	// Check for Alpine release file
	if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return true
	}

	// Check ldd output for musl - more reliable for detecting any musl-based system
	cmd := exec.Command("ldd", "--version")
	if output, err := cmd.CombinedOutput(); err == nil {
		if strings.Contains(strings.ToLower(string(output)), "musl") {
			return true
		}
	}

	return false
}

func libraryFilename() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "libturso_sync_sdk_kit.dylib", nil
	case "linux":
		if isMuslLibc() {
			return "libturso_sync_sdk_kit.a", nil
		} else {
			return "libturso_sync_sdk_kit.so", nil
		}
	case "windows":
		return "turso_sync_sdk_kit.dll", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func embeddedLibraryPath() (string, error) {
	if runtime.GOOS == "linux" && isMuslLibc() {
		return fmt.Sprintf("libs/%v_%v_musl", runtime.GOOS, runtime.GOARCH), nil
	}
	return fmt.Sprintf("libs/%v_%v", runtime.GOOS, runtime.GOARCH), nil
}

func embeddedLibraryHash() string {
	root, err := embeddedLibraryPath()
	if err != nil {
		return ""
	}
	filename, err := libraryFilename()
	if err != nil {
		return ""
	}
	hash, err := libs.ReadFile(fmt.Sprintf("%v.sha256", filepath.Join(root, filename)))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(hash))
}

func embeddedLibraryOpen() (fs.File, error) {
	root, err := embeddedLibraryPath()
	if err != nil {
		return nil, fmt.Errorf("can't create library path: %w", err)
	}
	filename, err := libraryFilename()
	if err != nil {
		return nil, fmt.Errorf("can't create library name: %w", err)
	}
	fullPath := filepath.Join(root, filename)
	file, err := libs.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("can't open embedded library: %w", err)
	}
	return file, nil
}

func embeddedLibraryTryCreate() (path string, err error) {
	cacheRoot := os.Getenv("TURSO_GO_CACHE_DIR")
	if cacheRoot == "" {
		if d, err := os.UserCacheDir(); err == nil {
			cacheRoot = d
		} else {
			cacheRoot = os.TempDir()
		}
	}
	hash := embeddedLibraryHash()
	// no embedded library, exit early
	if hash == "" {
		return "", nil
	}
	if len(hash) != 64 {
		return "", fmt.Errorf("invalid sha256 hash embedded to the loader binary: %v", hash)
	}
	filename, err := libraryFilename()
	if err != nil {
		return "", fmt.Errorf("inconsistency between supported platforms and embedded binaries: %w", err)
	}

	hashPath := hash[0:8]
	cacheDir := filepath.Join(cacheRoot, "turso-go", hashPath)

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache dir '%v' for library extraction: %w", cacheDir, err)
	}

	libraryPath := filepath.Join(cacheDir, filename)
	if file, err := os.Stat(libraryPath); err == nil && file.Size() > 0 {
		libraryData, err := os.Open(libraryPath)
		if err != nil {
			return "", fmt.Errorf("unable to read cached library file at %v to check hash: %w", libraryPath, err)
		}
		diskHash := sha256.New()
		_, err = io.Copy(diskHash, libraryData)
		if err != nil {
			return "", fmt.Errorf("unable to validate cached library file hash at %v: %w", libraryPath, err)
		}
		digest := fmt.Sprintf("%x", diskHash.Sum(nil))
		if digest != hash {
			return "", fmt.Errorf("cached library file hash sum mismatch at %v: expected=%v, got=%v", libraryPath, hash, digest)
		}
		return libraryPath, nil
	}

	library, err := embeddedLibraryOpen()
	if err != nil {
		return "", fmt.Errorf("failed to open embedded library: %w", err)
	}
	defer library.Close()

	embeddedHash := sha256.New()
	libraryHasher := io.TeeReader(library, embeddedHash)

	libraryCache, err := os.Create(libraryPath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file on disk at %v: %w", libraryPath, err)
	}
	defer libraryCache.Close()

	if _, err := io.Copy(libraryCache, libraryHasher); err != nil {
		return "", fmt.Errorf("failed to write library to the cache file at %v: %w", libraryPath, err)
	}
	if runtime.GOOS != "windows" {
		err = os.Chmod(libraryPath, 0o755)
		if err != nil {
			return "", fmt.Errorf("failed change library cache file permissions to 0755 at %v: %w", libraryPath, err)
		}
	}
	digest := fmt.Sprintf("%x", embeddedHash.Sum(nil))
	if digest != hash {
		return "", fmt.Errorf("embedded library file hash sum mismatch: expected=%v, got=%v", hash, digest)
	}
	return libraryPath, nil
}

func librarySystemSearch() (string, error) {
	filename, err := libraryFilename()
	if err != nil {
		return "", fmt.Errorf("failed to get library filename: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	switch runtime.GOOS {
	case "windows":
		paths := append(strings.Split(os.Getenv("PATH"), ";"), cwd)
		for _, path := range paths {
			dllPath := filepath.Join(path, filename)
			if _, err := os.Stat(dllPath); err == nil {
				return dllPath, nil
			}
		}
		return "", fmt.Errorf("library file %v not found at paths listed in PATH env var", filename)
	case "darwin", "linux":
		paths := append(strings.Split(os.Getenv("LD_LIBRARY_PATH"), ":"), cwd)
		for _, path := range paths {
			libPath := filepath.Join(path, filename)
			if _, err := os.Stat(libPath); err == nil {
				return libPath, nil
			}
		}
		return "", fmt.Errorf("library file %v not found at paths listed in LD_LIBRARY_PATH env var", filename)
	default:
		return "", fmt.Errorf("%v platform is not supported", runtime.GOOS)
	}
}
