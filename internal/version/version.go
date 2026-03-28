package version

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	assetVersion string
	once         sync.Once
)

// AssetVersion returns the current git commit SHA (from the working tree) for static asset cache busting.
// Resolved once on first use; falls back to "dev" if git is unavailable.
func AssetVersion() string {
	once.Do(func() {
		out, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			assetVersion = "dev"
			return
		}
		assetVersion = strings.TrimSpace(string(out))
		if assetVersion == "" {
			assetVersion = "dev"
		}
	})
	return assetVersion
}
