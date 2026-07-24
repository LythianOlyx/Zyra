package tailwind

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// parseSHA256Sums parses standard sha256sum output format:
// lines of `<hex-digest>  ./<filename>` or `<hex-digest>  <filename>`.
func parseSHA256Sums(content string) map[string]string {
	sums := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			hash := strings.ToLower(fields[0])
			if len(hash) == 64 {
				filename := fields[1]
				filename = strings.TrimPrefix(filename, "./")
				sums[filename] = hash
			}
		}
	}
	return sums
}

// verifyChecksum verifies downloaded binary data against explicitChecksum (if configured)
// and/or automatically against sha256sums.txt downloaded from the release.
func verifyChecksum(ctx context.Context, client *http.Client, data []byte, assetName, version, downloadURLOverride, explicitChecksum string, logger *zap.Logger) error {
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])

	if explicitChecksum != "" {
		if !strings.EqualFold(actual, explicitChecksum) {
			return fmt.Errorf("zyra/render/tailwind: checksum mismatch for asset %s: expected %s, got %s", assetName, explicitChecksum, actual)
		}
	}

	sumsURL := buildDownloadURL(version, downloadURLOverride, "sha256sums.txt")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumsURL, nil)
	if err != nil {
		if explicitChecksum == "" {
			logger.Warn("zyra/render/tailwind: checksum verification skipped because sha256sums.txt request build failed", zap.Error(err))
		}
		return nil
	}

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		if explicitChecksum == "" {
			logger.Warn("zyra/render/tailwind: checksum verification skipped because sha256sums.txt unavailable", zap.String("url", sumsURL), zap.Error(err))
		}
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if explicitChecksum == "" {
			logger.Warn("zyra/render/tailwind: checksum verification skipped because sha256sums.txt returned non-200 status", zap.String("url", sumsURL), zap.Int("status", resp.StatusCode))
		}
		return nil
	}

	sumsBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if explicitChecksum == "" {
			logger.Warn("zyra/render/tailwind: checksum verification skipped because reading sha256sums.txt failed", zap.Error(err))
		}
		return nil
	}

	sumsMap := parseSHA256Sums(string(sumsBody))
	expectedHash, ok := sumsMap[assetName]
	if !ok {
		if explicitChecksum == "" {
			logger.Warn("zyra/render/tailwind: asset not found in sha256sums.txt, checksum verification skipped", zap.String("asset", assetName))
		}
		return nil
	}

	if !strings.EqualFold(actual, expectedHash) {
		return fmt.Errorf("zyra/render/tailwind: checksum mismatch for asset %s against sha256sums.txt: expected %s, got %s", assetName, expectedHash, actual)
	}

	return nil
}
