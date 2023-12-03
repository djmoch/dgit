// See LICENSE file for copyright and license details

package main

import (
	"bytes"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(0)

	fix := flag.Bool("fix", false, "fix files in addition to printing their paths")
	flag.Parse()

	cmdArgs := []string{"-l"}
	if *fix {
		cmdArgs = append(cmdArgs, "-w")
	}
	cmdArgs = append(cmdArgs, flag.Args()...)
	cmd := exec.Command("goimports", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			log.Print(string(exitErr.Stderr))
		}
		log.Fatal("goimports failed: ", err)
	}
	if len(output) != 0 {
		files := bytes.Split(output, []byte{'\n'})
		for _, file := range files {
			if len(file) > 0 {
				if *fix {
					log.Printf("%s: fixed imports and format", string(file))
				} else {
					log.Printf("%s: run goimports", string(file))
				}
			}
		}
		os.Exit(1)
	}
}
