package tailwind

import (
	"fmt"
	"os"
	"strings"
)

// defaultFileExists checks if a file exists on the host filesystem.
func defaultFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isMusl reports whether the environment requires a musl-linked binary.
// It prioritizes explicit libcConfig ("musl" vs "glibc"), falling back to
// checking for /etc/alpine-release via hasFile when libcConfig is empty.
func isMusl(libcConfig string, hasFile func(string) bool) bool {
	switch strings.ToLower(libcConfig) {
	case "musl":
		return true
	case "glibc":
		return false
	default:
		if hasFile != nil {
			return hasFile("/etc/alpine-release")
		}
		return defaultFileExists("/etc/alpine-release")
	}
}

// parseMajorVersion extracts the major version integer from a version string (e.g. "4.3.3" -> 4).
func parseMajorVersion(v string) int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return 0
	}
	var major int
	_, err := fmt.Sscanf(parts[0], "%d", &major)
	if err != nil {
		return 0
	}
	return major
}

// platformAsset returns the official Tailwind CSS Standalone CLI release asset
// name for the given GOOS/GOARCH pair, assuming default version and auto-detected libc.
// Retained for backward compatibility.
func platformAsset(goos, goarch string) (string, error) {
	return platformAssetExtended(goos, goarch, "", "", nil)
}

// platformAssetExtended returns the exact official Tailwind CSS Standalone CLI
// release asset name for the given GOOS/GOARCH pair, version, and libc environment.
//
// It is a pure function of its arguments — taking version, libcConfig, and a hasFile callback —
// so every platform, version, and libc combination can be exercised from unit tests regardless
// of the machine actually running them.
//
// Tailwind v4+ publishes musl-linked binaries (tailwindcss-linux-x64-musl, tailwindcss-linux-arm64-musl)
// for Alpine Linux support. Tailwind v3 releases do not publish musl binaries.
func platformAssetExtended(goos, goarch, version, libcConfig string, hasFile func(string) bool) (string, error) {
	major := parseMajorVersion(version)
	musl := isMusl(libcConfig, hasFile)

	switch {
	case goos == "linux" && goarch == "amd64":
		if major >= 4 && musl {
			return "tailwindcss-linux-x64-musl", nil
		}
		return "tailwindcss-linux-x64", nil
	case goos == "linux" && goarch == "arm64":
		if major >= 4 && musl {
			return "tailwindcss-linux-arm64-musl", nil
		}
		return "tailwindcss-linux-arm64", nil
	case goos == "darwin" && goarch == "amd64":
		return "tailwindcss-macos-x64", nil
	case goos == "darwin" && goarch == "arm64":
		return "tailwindcss-macos-arm64", nil
	case goos == "windows" && goarch == "amd64":
		return "tailwindcss-windows-x64.exe", nil
	default:
		return "", fmt.Errorf("zyra/render/tailwind: unsupported platform GOOS=%q GOARCH=%q: the official Tailwind CSS Standalone CLI does not publish a binary for this combination (set TailwindConfig.BinaryPath to use a manually provisioned binary instead)", goos, goarch)
	}
}

// buildDownloadURL returns the URL the given platform asset should be
// downloaded from: the official GitHub Releases URL for version, unless
// override is non-empty, in which case override is treated as an
// enterprise/air-gapped mirror base URL and asset is appended to it —
// mirroring the documented behavior of TailwindConfig.DownloadURL.
func buildDownloadURL(version, override, asset string) string {
	if override != "" {
		return strings.TrimSuffix(override, "/") + "/" + asset
	}
	return fmt.Sprintf("https://github.com/tailwindlabs/tailwindcss/releases/download/v%s/%s", version, asset)
}
