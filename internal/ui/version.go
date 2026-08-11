package ui

import (
	"runtime/debug"
	"strings"
)

// repoURL is the fallback project link used when the module path can't be read
// from the build info (e.g. a plain `go build` in some environments).
const repoURL = "https://github.com/cwqt/sshblog"

// buildInfo returns a display version ("0.2.1", or "dev" for an untagged build)
// and the project's repository URL, both derived from the compiled module's
// build info. Installing a tagged release (`go install …@v0.2.1`) yields the
// real version; a local `go build` yields "dev".
func buildInfo() (version, url string) {
	version, url = "dev", repoURL
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		version = strings.TrimPrefix(v, "v")
	}
	if p := info.Main.Path; p != "" {
		url = "https://" + p
	}
	return
}

// hyperlink wraps text in an OSC 8 escape so supporting terminals render it as
// a clickable link to url; terminals without OSC 8 just show the plain text.
func hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}
