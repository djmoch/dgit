// See LICENSE file for copyright and license details

package assets

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

var minify = flag.Bool("minify", false, "if true, re-write minified JS/CSS")

type minifyRequest struct {
	in, out string
	t       *testing.T
}

func TestJSUpToDate(t *testing.T) {
	req := minifyRequest{
		in:  "highlight.js",
		out: "highlight.min.js",
		t:   t,
	}
	doMinify(req)
}

func TestCSSUpToDate(t *testing.T) {
	req := minifyRequest{
		in:  "site.css",
		out: "site.min.css",
		t:   t,
	}
	doMinify(req)
}

func doMinify(req minifyRequest) {
	r := api.Build(api.BuildOptions{
		EntryPoints:       []string{req.in},
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Outdir:            ".",
		OutExtension:      map[string]string{".js": ".min.js", ".css": ".min.css"},
		Sourcemap:         api.SourceMapLinked,
	})
	if len(r.Errors) != 0 || len(r.Warnings) != 0 {
		for _, msg := range r.Errors {
			if msg.Location != nil {
				req.t.Errorf("%s:%d:%d %s", req.in, msg.Location.Line,
					msg.Location.Column, msg.Text)
			} else {
				req.t.Error(msg.Text)
			}
		}
		return
	}

	old, err := os.ReadFile(req.out)
	if err != nil {
		req.t.Logf("Failed to read %s. Assuming it doesn't exist.", req.out)
		old = []byte("")
	}

	for _, f := range r.OutputFiles {
		if filepath.Base(f.Path) == req.out {
			if string(f.Contents) == string(old) {
				req.t.Log(req.out, " up to date")
				return
			}
		}
	}

	if *minify {
		for _, f := range r.OutputFiles {
			if err := os.WriteFile(f.Path, f.Contents, 0666); err != nil {
				req.t.Fatal(err)
			}
			req.t.Logf("write %d bytes to %s", len(f.Contents), f.Path)
		}
	} else {
		req.t.Error(req.out, "stale. To update, run 'go generate ./cmd/dgit/internal/serve/assets'")
	}
}
