// See LICENSE file for copyright and license details

// Package version implements the "dgit version" command
package version

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"runtime/debug"

	"djmo.ch/dgit/cmd/dgit/internal/base"
)

var (
	version = "0.1.0"

	// Taken from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	reString = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	re       = regexp.MustCompile(reString)
)

var Cmd = &base.Command{
	Run:       printVersion,
	Usage:     "version",
	Name:      "version",
	ShortHelp: "print version information",
	LongHelp: `Version prints program version information.

Version information is compliant with version 2.0.0 of the semver
spec. If the version includes a pre-release identifier, build metadata
is also appended to the reported version.
	`,
}

func printVersion(ctx context.Context) {
	var (
		w = ctx.Value("w").(io.Writer)

		buildmetadata, rev, dirty string
	)
	if !re.MatchString(version) {
		panic("version is not semver-compliant!")
	}
	if re.FindStringSubmatch(version)[re.SubexpIndex("prerelease")] == "" {
		fmt.Fprintln(w, "dgit version", version)
		return
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintln(w, "dgit version", version+"+nobuildinfo")
		return
	}
	buildmetadata += "+"
	for _, bs := range info.Settings {
		switch bs.Key {
		case "vcs.revision":
			rev += bs.Value[:7]
		case "vcs.modified":
			if bs.Value == "true" {
				dirty += ".dirty"
			}
		}
	}
	buildmetadata += rev + dirty
	fmt.Fprintln(w, "dgit version", version+buildmetadata)
}
