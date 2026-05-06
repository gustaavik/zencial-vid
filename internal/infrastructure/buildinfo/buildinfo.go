package buildinfo

// Version is updated by the Release workflow on each published tag.
// Commit and BuildTime are set at build time via -ldflags.
var (
	Version   = "v0.2.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)
