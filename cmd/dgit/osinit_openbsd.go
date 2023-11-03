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
			panic(fmt.Sprint("filepath.Abs: ", err))
		}
		projectListPath, err := filepath.Abs(cfg.ProjectListPath)
		if err != nil {
			panic(fmt.Sprintf("filepath.Abs: ", err))
		}
		err = unix.Unveil(repoBasePath, "r")
		if err != nil {
			panic(fmt.Sprint("unix.Unveil: ", err))
		}
		err = unix.Unveil(projectListPath, "r")
		if err != nil {
			panic(fmt.Sprint("unix.Unveil: ", err))
		}
		err = unix.Pledge("stdio rpath dns inet flock", "")
		if err != nil {
			panic(fmt.Sprint("unix.Pledge: ", err))
		}
	}
}
