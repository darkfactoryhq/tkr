package main

import (
	"os"
	"runtime/debug"

	"github.com/darkfactoryhq/tkr/internal/cmd"
)

// These are overridden at release time by GoReleaser via -ldflags -X.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(resolveVersion())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveVersion returns the version, commit, and build date. GoReleaser injects
// these via ldflags for released binaries. For binaries built with
// `go install <module>@<version>` — where those ldflags are not applied — fall
// back to the module version and VCS metadata embedded by the Go toolchain so
// the binary still reports a meaningful version instead of "dev".
func resolveVersion() (v, c, d string) {
	v, c, d = version, commit, date
	if v != "dev" {
		return v, c, d
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v, c, d
	}

	if mv := info.Main.Version; mv != "" && mv != "(devel)" {
		v = mv
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			c = s.Value
			if len(c) > 7 {
				c = c[:7]
			}
		case "vcs.time":
			d = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				c += "-dirty"
			}
		}
	}
	return v, c, d
}
