package config

import (
	"os"
	"strings"
)

// ResolveSecret resolves a secret value using the following priority:
//  1. Contents of the file pointed to by <envKey>_FILE (Docker/Swarm secret)
//  2. The environment variable <envKey>
//  3. The fallback value (typically read from the config YAML)
func ResolveSecret(envKey string, fallback string) string {
	if filePath := os.Getenv(envKey + "_FILE"); filePath != "" {
		if b, err := os.ReadFile(filePath); err == nil {
			if v := strings.TrimSpace(string(b)); v != "" {
				return v
			}
		}
	}
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		return v
	}
	return fallback
}
