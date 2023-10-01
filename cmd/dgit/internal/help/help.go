// See LICENSE file for copyright and license details

// Package help implements the "dgit help" command
package help

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"

	"djmo.ch/dgit/cmd/dgit/internal/base"
)

var usageTmpl = `{{ if .Usage }}usage: {{ .Usage }}

{{ end }}{{ .LongHelp | trim }}{{ if .Subcommands }}

The commands are:
{{ range .Subcommands -}}{{ if .Runnable }}
	{{ .Name | printf "%-8s"}}{{ .ShortHelp }}{{ end }}{{ end }}

Use "dgit help <command>" for more information about a command.

Additional topics are:
{{ range .Subcommands -}}{{ if (not .Runnable) }}
	{{ .Name | printf "%-12s"}}{{ .ShortHelp }}{{ end }}{{ end }}

Use "dgit help <topic>" for more information about that topic.{{ end }}
`

func Help(ctx context.Context) {
	var (
		args = ctx.Value("args").([]string)
		w    = ctx.Value("w").(io.Writer)
	)
	if args == nil {
		args = []string{}
	}
	switch len(args) {
	case 0:
		printUsage(w, base.DGit)
	case 1:
		if args[0] == "documentation" {
			printDocumentation(w)
			return
		}
		cmd := base.FindCommand(args[0])
		if cmd == nil {
			log.Fatal("unknown subcommand ", args[0])
		}
		printUsage(w, cmd)
	default:
		log.Fatal("help expects at most one argument")
	}
}

func printDocumentation(w io.Writer) {
	fmt.Fprintln(w, "// See LICENSE file for copyright and license details")
	fmt.Fprintln(w, "// Code generated by 'go test ./cmd/dgit -v -run=TestDocsUpToDate -fixdocs'; DO NOT EDIT")
	fmt.Fprintln(w, "// Edit the code in other files and then execute 'go generate ./cmd/dgit' to generate this one.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "/*")
	printUsage(w, base.DGit)
	for _, cmd := range base.DGit.Subcommands {
		if cmd.LongHelp != "" {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "# %s%s\n", strings.ToTitle(string(cmd.ShortHelp[0])),
				cmd.ShortHelp[1:])
			fmt.Fprintln(w)
			printUsage(w, cmd)
		}
	}
	fmt.Fprintln(w, "*/")
	fmt.Fprintln(w, "package main")
}

func printUsage(w io.Writer, cmd *base.Command) {
	tmpl := template.New(cmd.Name)
	tmpl.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(tmpl.Parse(usageTmpl))
	tmpl.Execute(w, cmd)
}
