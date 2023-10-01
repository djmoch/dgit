// See LICENSE file for copyright and license details

package base

import (
	"context"
	"testing"
)

func TestFindCommand(t *testing.T) {
	cmd := &Command{Name: "test"}
	DGit.Subcommands = []*Command{cmd}
	ret := FindCommand("test")
	if ret == nil {
		t.Error("FindCommand did not find 'test' command")
	}
	ret = FindCommand("foo")
	if ret != nil {
		t.Error("FindCommand found non-existant command")
	}
}

func TestRunnable(t *testing.T) {
	f := func(context.Context) {}
	cmd := &Command{
		Name: "test",
		Run:  f,
	}
	cmd2 := &Command{Name: "test2"}
	if !cmd.Runnable() {
		t.Error("command expected to be runnable")
	}
	if cmd2.Runnable() {
		t.Error("command expected not to be runnable")
	}
}
