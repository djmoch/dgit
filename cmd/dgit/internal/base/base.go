// See LICENSE file for copyright and license details

// Package base defines the foundational structures required to build
// out the dgit command suite.
package base

import (
	"context"
	"flag"
)

// Environment variable keys
const (
	DGITENV             = "DGITENV"
	DGIT_REPO_BASE      = "DGIT_REPO_BASE"
	DGIT_PROJ_LIST_PATH = "DGIT_PROJ_LIST_PATH"
	DGIT_REMOVE_SUFFIX  = "DGIT_REMOVE_SUFFIX"
)

// KnownEnv is a list of environment variables that affect the
// operation of the dgit command
const KnownEnv = `
	DGITENV
	DGIT_REPO_BASE
	DGIT_PROJ_LIST_PATH
	DGIT_REMOVE_SUFFIX
	`

type Command struct {
	Run                              func(context.Context)
	Flags                            flag.FlagSet
	Name, ShortHelp, LongHelp, Usage string
	Subcommands                      []*Command
}

var DGit = &Command{
	Name: "dgit",
	LongHelp: `Djmoch's Git Viewer

DGit is a template-driven alternative to CGit. It runs either as a
FastCGI utility or listens on a TCP port to allow viewing Git
repositories from a web browser. The look-and-feel of the website is
controlled by templates. You can use the provided templates, or you
can roll your own.
`,
	Usage: "dgit <command> [arguments]",
}

func FindCommand(cmd string) *Command {
	for _, sub := range DGit.Subcommands {
		if sub.Name == cmd {
			return sub
		}
	}
	return nil
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}
