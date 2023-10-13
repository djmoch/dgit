// See LICENSE file for copyright and license details

package main

import (
	"fmt"
	"path/filepath"

	"djmo.ch/dgit/config"
	"golang.org/x/sys/unix"
)

func init() {
	osInit = func(cfg config.Config) {
		repoBasePath, err := filepath.Abs(cfg.RepoBasePath)
		if err != nil {
			panic(fmt.Sprint("RepoBasePath could not be made absolute:", err))
		}
		projectListPath, err := filepath.Abs(cfg.ProjectListPath)
		if err != nil {
			panic(fmt.Sprintf("ProjectListPath could not be made absolute:", err))
		}
		err = unix.Unveil(repoBasePath, "r")
		if err != nil {
			panic(fmt.Sprint("could not unveil RepoBasePath:", err))
		}
		err = unix.Unveil(projectListPath, "r")
		if err != nil {
			panic(fmt.Sprint("could not unveil ProjectListPath:", err))
		}
		err = unix.Pledge("stdio rpath dns inet flock", "")
		if err != nil {
			panic(fmt.Sprint("pledge failed:", err))
		}
	}
}
