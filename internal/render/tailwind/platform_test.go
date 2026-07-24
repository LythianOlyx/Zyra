package tailwind

import (
	"strings"
	"testing"
)

func TestPlatformAsset_SupportedCombinations(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "tailwindcss-linux-x64"},
		{"linux", "arm64", "tailwindcss-linux-arm64"},
		{"darwin", "amd64", "tailwindcss-macos-x64"},
		{"darwin", "arm64", "tailwindcss-macos-arm64"},
		{"windows", "amd64", "tailwindcss-windows-x64.exe"},
	}

	for _, c := range cases {
		t.Run(c.goos+"_"+c.goarch, func(t *testing.T) {
			got, err := platformAsset(c.goos, c.goarch)
			if err != nil {
				t.Fatalf("platformAsset(%q, %q) returned unexpected error: %v", c.goos, c.goarch, err)
			}
			if got != c.want {
				t.Errorf("platformAsset(%q, %q) = %q, want %q", c.goos, c.goarch, got, c.want)
			}
		})
	}
}

func TestPlatformAssetExtended_MuslAndV4Matrix(t *testing.T) {
	fakeAlpine := func(path string) bool {
		return path == "/etc/alpine-release"
	}
	fakeNonAlpine := func(path string) bool {
		return false
	}

	cases := []struct {
		name    string
		goos    string
		goarch  string
		version string
		libc    string
		hasFile func(string) bool
		want    string
		wantErr bool
	}{
		{
			name:    "linux amd64 v4 musl explicit",
			goos:    "linux",
			goarch:  "amd64",
			version: "4.3.3",
			libc:    "musl",
			hasFile: fakeNonAlpine,
			want:    "tailwindcss-linux-x64-musl",
		},
		{
			name:    "linux arm64 v4 musl explicit",
			goos:    "linux",
			goarch:  "arm64",
			version: "4.3.3",
			libc:    "musl",
			hasFile: fakeNonAlpine,
			want:    "tailwindcss-linux-arm64-musl",
		},
		{
			name:    "linux amd64 v4 alpine auto-detect",
			goos:    "linux",
			goarch:  "amd64",
			version: "4.3.3",
			libc:    "",
			hasFile: fakeAlpine,
			want:    "tailwindcss-linux-x64-musl",
		},
		{
			name:    "linux amd64 v4 glibc explicit on alpine",
			goos:    "linux",
			goarch:  "amd64",
			version: "4.3.3",
			libc:    "glibc",
			hasFile: fakeAlpine,
			want:    "tailwindcss-linux-x64",
		},
		{
			name:    "linux amd64 v3 musl explicit stays glibc (v3 has no musl assets)",
			goos:    "linux",
			goarch:  "amd64",
			version: "3.4.17",
			libc:    "musl",
			hasFile: fakeAlpine,
			want:    "tailwindcss-linux-x64",
		},
		{
			name:    "darwin arm64 ignore libc",
			goos:    "darwin",
			goarch:  "arm64",
			version: "4.3.3",
			libc:    "musl",
			hasFile: fakeAlpine,
			want:    "tailwindcss-macos-arm64",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := platformAssetExtended(c.goos, c.goarch, c.version, c.libc, c.hasFile)
			if (err != nil) != c.wantErr {
				t.Fatalf("platformAssetExtended error = %v, wantErr = %v", err, c.wantErr)
			}
			if got != c.want {
				t.Errorf("platformAssetExtended = %q, want %q", got, c.want)
			}
		})
	}
}

func TestPlatformAsset_UnsupportedCombinations(t *testing.T) {
	cases := []struct {
		goos, goarch string
	}{
		{"linux", "386"},
		{"linux", "arm"},
		{"linux", "riscv64"},
		{"windows", "arm64"},
		{"windows", "386"},
		{"darwin", "arm"},
		{"darwin", "386"},
		{"freebsd", "amd64"},
		{"openbsd", "amd64"},
		{"js", "wasm"},
		{"plan9", "amd64"},
		{"solaris", "amd64"},
		{"", ""},
		{"linux", ""},
		{"", "amd64"},
	}

	for _, c := range cases {
		t.Run(c.goos+"_"+c.goarch, func(t *testing.T) {
			got, err := platformAsset(c.goos, c.goarch)
			if err == nil {
				t.Fatalf("platformAsset(%q, %q) = %q, want an error", c.goos, c.goarch, got)
			}
			if got != "" {
				t.Errorf("platformAsset(%q, %q) returned non-empty asset %q alongside an error", c.goos, c.goarch, got)
			}
			if c.goos != "" && !strings.Contains(err.Error(), c.goos) {
				t.Errorf("expected error to mention GOOS %q, got: %v", c.goos, err)
			}
			if c.goarch != "" && !strings.Contains(err.Error(), c.goarch) {
				t.Errorf("expected error to mention GOARCH %q, got: %v", c.goarch, err)
			}
		})
	}
}

func TestBuildDownloadURL_DefaultUsesOfficialGitHubReleases(t *testing.T) {
	got := buildDownloadURL("3.4.17", "", "tailwindcss-linux-x64")
	want := "https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64"
	if got != want {
		t.Errorf("buildDownloadURL() = %q, want %q", got, want)
	}
}

func TestBuildDownloadURL_MirrorOverrideAppendsAssetName(t *testing.T) {
	cases := []struct {
		override string
		want     string
	}{
		{"https://mirror.internal/tailwind", "https://mirror.internal/tailwind/tailwindcss-linux-x64"},
		{"https://mirror.internal/tailwind/", "https://mirror.internal/tailwind/tailwindcss-linux-x64"},
	}

	for _, c := range cases {
		got := buildDownloadURL("3.4.17", c.override, "tailwindcss-linux-x64")
		if got != c.want {
			t.Errorf("buildDownloadURL(override=%q) = %q, want %q", c.override, got, c.want)
		}
	}
}
