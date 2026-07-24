package tailwind

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ResolveLatestVersion queries the GitHub API for the latest Tailwind CSS release
// and returns the tag version without the leading 'v' (e.g. "4.3.3").
//
// It uses client (or http.DefaultClient if nil), sets a User-Agent header as required
// by GitHub's API, and respects explicit bearer tokens or GITHUB_TOKEN/GH_TOKEN environment
// variables for authenticated requests. Non-200 responses and 403 rate limits return
// actionable errors.
func ResolveLatestVersion(ctx context.Context, client *http.Client, token string) (string, error) {
	return ResolveLatestVersionFromURL(ctx, client, token, "https://api.github.com/repos/tailwindlabs/tailwindcss/releases/latest")
}

// ResolveLatestVersionFromURL is an internal variant allowing custom endpoint URLs in tests.
func ResolveLatestVersionFromURL(ctx context.Context, client *http.Client, token, endpoint string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("zyra/render/tailwind: failed to create request for latest version: %w", err)
	}

	req.Header.Set("User-Agent", "Zyra-Framework-Tailwind-Resolver/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GH_TOKEN")
		}
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("zyra/render/tailwind: failed to query GitHub API for latest release at %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return "", fmt.Errorf("zyra/render/tailwind: GitHub API rate limit exceeded (60 requests/hour for unauthenticated calls). Set GITHUB_TOKEN environment variable or configure TailwindConfig.GitHubToken to authenticate")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("zyra/render/tailwind: failed to resolve latest tailwind version from GitHub API at %s: unexpected HTTP status %d %s", endpoint, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("zyra/render/tailwind: failed to read GitHub API response: %w", err)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("zyra/render/tailwind: failed to parse GitHub API release JSON: %w", err)
	}

	if payload.TagName == "" {
		return "", fmt.Errorf("zyra/render/tailwind: GitHub API response missing tag_name field")
	}

	version := strings.TrimPrefix(payload.TagName, "v")
	return version, nil
}
