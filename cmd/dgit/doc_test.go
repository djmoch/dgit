// See LICENSE file for copyright and license details

package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"djmo.ch/dgit/cmd/dgit/internal/help"
)

var fixdocs = flag.Bool("fixdocs", false, "if true, update doc.go")

func TestDocsUpToDate(t *testing.T) {
	var (
		w   = new(strings.Builder)
		ctx = context.Background()
	)
	ctx = context.WithValue(ctx, "w", w)
	ctx = context.WithValue(ctx, "args", []string{"documentation"})
	help.Help(ctx)
	newfile := w.String()
	old, err := os.ReadFile("doc.go")
	if err != nil {
		t.Log("Failed to read doc.go. Assuming it doesn't exist.")
		old = []byte("")
	}
	if newfile == string(old) {
		t.Log("doc.go up to date")
		return
	}

	if *fixdocs {
		if err := os.WriteFile("doc.go", []byte(newfile), 0666); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d bytes to doc.go", len([]byte(newfile)))
	} else {
		t.Error("doc.go stale. To update, run 'go generate ./cmd/dgit'")
	}
}
