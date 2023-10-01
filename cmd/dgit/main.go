// See LICENSE file for copyright and license details

//go:generate go test -v -run=TestDocsUpToDate -fixdocs

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"djmo.ch/dgit/cmd/dgit/internal/base"
	"djmo.ch/dgit/cmd/dgit/internal/env"
	"djmo.ch/dgit/cmd/dgit/internal/help"
	"djmo.ch/dgit/cmd/dgit/internal/serve"
	"djmo.ch/dgit/cmd/dgit/internal/version"
)

var osInit func()

func init() {
	base.DGit.Subcommands = []*base.Command{
		serve.Cmd,
		env.Cmd,
		version.Cmd,

		help.EnvCmd,
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")
	flag.Usage = usage
	flag.Parse()

	if osInit != nil {
		osInit()
	}

	env.MergeEnv()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "args", args[1:])
	ctx = context.WithValue(ctx, "w", os.Stdout)
	ctx = context.WithValue(ctx, "cfg", env.ConfigFromEnv())

	if args[0] == "help" {
		help.Help(ctx)
		return
	}

	cmd := base.FindCommand(args[0])
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "%s %s: unknown command\n", os.Args[0], args[0])
		fmt.Fprintf(os.Stderr, "Run '%s help' for usage\n", os.Args[0])
		os.Exit(1)
	}

	cmd.Flags.Parse(os.Args[2:])
	ctx = context.WithValue(ctx, "args", cmd.Flags.Args())

	cmd.Run(ctx)
}

func usage() {
	var (
		ctx  = context.Background()
		args = make([]string, 0, 0)
	)
	ctx = context.WithValue(ctx, "w", os.Stdout)
	ctx = context.WithValue(ctx, "args", args)
	help.Help(ctx)
}
