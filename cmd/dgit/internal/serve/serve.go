// See LICENSE file for copyright and license details

// Package serve implements the "dgit serve" command
package serve

import (
	"context"
	"embed"
	"log"
	"net"
	"net/http"
	"net/url"

	"djmo.ch/dgit"
	"djmo.ch/dgit/cmd/dgit/internal/base"
	"djmo.ch/dgit/config"
)

var Cmd = &base.Command{
	Run:       runServe,
	Name:      "serve",
	Usage:     "dgit serve URL",
	ShortHelp: "serve Git repositories",
	LongHelp: `Serve serves Git repositories

DGit listens and serves repositories on the provided URL. The only
recognized scheme is http.

The DGit handler supports the "dumb" Git HTTP protocol, so read-only
repository operations, such as cloning and fetching, are supported.
	`,
}

//go:embed assets/*
var assets embed.FS

func runServe(ctx context.Context) {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")
	var (
		args = ctx.Value("args").([]string)
		cfg  = ctx.Value("cfg").(config.Config)
	)
	if len(args) != 1 {
		log.Fatal("no URL provided")
	}
	u, err := url.Parse(args[0])
	if err != nil {
		log.Fatal("failed to parse URL: ", err)
	}
	dg := &dgit.DGit{Config: cfg}
	http.Handle("/", dg)
	http.Handle("/-/", http.StripPrefix("/-/", http.FileServer(http.FS(assets))))
	switch u.Scheme {
	case "http":
		listener, err := net.Listen("tcp", u.Host)
		if err != nil {
			log.Fatal("listen: ", err)
		}
		log.Fatal(http.Serve(listener, nil))
	default:
		log.Fatal("unknown scheme:", u.Scheme)
	}
}
