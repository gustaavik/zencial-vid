package buildinfo

import "os"

// Set at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func init() {
	sc := os.Getenv("SOURCE_COMMIT")
	if sc == "" {
		return
	}

	if Commit == "unknown" {
		if len(sc) > 7 {
			Commit = sc[:7]
		} else {
			Commit = sc
		}
	}

	if Version == "dev" || Version == "0.0.0-unknown" {
		Version = "0.0.0-" + Commit
	}
}
