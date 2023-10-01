// See LICENSE file for copyright and license details

package help

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"djmo.ch/dgit/cmd/dgit/internal/base"
)

func TestHelp(t *testing.T) {
	var (
		s   = "Test is a fake command"
		cmd = &base.Command{
			Name:     "test",
			LongHelp: s,
		}
		w    = new(strings.Builder)
		ctx  = context.Background()
		args = make([]string, 0, 1)
	)
	base.DGit.Subcommands = []*base.Command{cmd}
	ctx = context.WithValue(ctx, "w", w)
	ctx = context.WithValue(ctx, "args", args)
	fmt.Print(ctx.Value("args"))
	Help(ctx)
	if !strings.HasPrefix(w.String(), base.DGit.ShortHelp) {
		t.Error("Unexpected dgit help text")
	}
	w.Reset()
	args = append(args, "test")
	ctx = context.WithValue(ctx, "args", args)
	Help(ctx)
	if !strings.HasPrefix(w.String(), s) {
		t.Error("Unexpected subcommand help text")
	}
}
