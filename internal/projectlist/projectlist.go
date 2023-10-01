// See LICENSE file for copyright and license details

package projectlist

import (
	"bufio"
	"fmt"
	"os"
)

// A ProjectList contains a list of repositories according to their
// filesystem path. When a repository is not bare, its path is
// considered to be the path to the "git directory" (usually the .git
// directory within the main worktree.
type ProjectList []string

func NewProjectList(listPath string) (ProjectList, error) {
	var pl ProjectList
	if listPath == "" {
		return pl, fmt.Errorf("NewProjectList: no path to project list specified")
	}
	listFile, err := os.Open(listPath)
	if err != nil {
		return pl,
			fmt.Errorf("newProjectList: failure opening project list file: %s",
				err)
	}
	defer listFile.Close()
	scanner := bufio.NewScanner(listFile)
	for scanner.Scan() {
		pl = append(pl, scanner.Text())
	}
	return pl, nil
}
