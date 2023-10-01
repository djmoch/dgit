// See LICENSE file for copyright and license details

// Package serve implements the "dgit serve" command
package serve

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
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

Repositories are served for viewing on the provided URL. When the URL
specifies a Unix domain socket, DGit acts as an FCGI server, otherwise
it acts as an HTTP server. DGit does not serve over HTTPS, and as such
specifying that scheme is an error.
	`,
}

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
	if u.Scheme == "https" {
		log.Fatal("https scheme not supported")
	}
	dg := &dgit.DGit{Config: cfg}
	switch u.Scheme {
	case "http":
		listener, err := net.Listen("tcp", u.Host)
		if err != nil {
			log.Fatal("listen: ", err)
		}
		log.Fatal(http.Serve(listener, dg))
	case "unix":
		listener, err := net.Listen("unix", u.Path)
		if err != nil {
			log.Fatal("listen: ", err)
		}
		log.Fatal(fcgi.Serve(listener, dg))
	default:
		log.Fatal("unknown scheme:", u.Scheme)
	}
}
