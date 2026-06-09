package version

// Variables injected at build time via -ldflags.
var (
	// Version is the semantic version (e.g. "v1.0.0").
	// Default: "devel".
	Version = "devel"

	// Commit is the git commit hash.
	// Default: "unknown".
	Commit = "unknown"

	// Date is the build timestamp.
	// Default: "unknown".
	Date = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return "v" + Version + " (commit " + Commit + ", built " + Date + ")"
}
