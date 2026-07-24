package bundler

import (
	"encoding/json"
	"os"
	"strings"
)

// PublicEnvPrefix is the required prefix for environment variables accessible in client bundles.
const PublicEnvPrefix = "PUBLIC_"

// FilterPublicEnv takes a map of environment variables and returns an esbuild Define map.
// Only variables starting with "PUBLIC_" (or explicit system vars like NODE_ENV) are included.
// Any non-PUBLIC_ environment variables passed in or detected in os.Environ() are mapped to "undefined"
// to prevent sensitive keys (e.g., DATABASE_URL, STRIPE_SECRET_KEY) from reaching client bundles.
func FilterPublicEnv(envVars map[string]string) map[string]string {
	define := make(map[string]string)

	// Combine passed envVars with system environment
	allVars := make(map[string]string)
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			allVars[parts[0]] = parts[1]
		}
	}
	for k, v := range envVars {
		allVars[k] = v
	}

	for k, v := range allVars {
		if strings.HasPrefix(k, PublicEnvPrefix) {
			jsonVal, err := json.Marshal(v)
			if err != nil {
				continue
			}
			valStr := string(jsonVal)
			define["process.env."+k] = valStr
			define["import.meta.env."+k] = valStr
		} else if k != "PATH" && k != "HOME" && k != "SHELL" && k != "USER" && k != "TMPDIR" {
			// Explicitly replace non-PUBLIC_ variable accesses with undefined
			define["process.env."+k] = "undefined"
			define["import.meta.env."+k] = "undefined"
		}
	}

	// Always set process.env.NODE_ENV if present in envVars or default to development
	nodeEnv := allVars["NODE_ENV"]
	if nodeEnv == "" {
		nodeEnv = allVars["ZYRA_ENV"]
	}
	if nodeEnv == "" {
		nodeEnv = "development"
	}
	jsonNodeEnv, _ := json.Marshal(nodeEnv)
	define["process.env.NODE_ENV"] = string(jsonNodeEnv)
	define["import.meta.env.NODE_ENV"] = string(jsonNodeEnv)

	return define
}
