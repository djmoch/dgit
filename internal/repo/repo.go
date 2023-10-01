// See LICENSE file for copyright and license details

package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"djmo.ch/dgit/config"
	git "github.com/go-git/go-git/v5"
)

const lastModifiedFormat = "2006-01-02 15:04:05 -0700"

// A repo contains information for a single repository.
type Repo struct {
	// Path is the repository path relative to RepoBathPath.
	Path string
	// Slug is the URL path to the repository, relative to the
	// DGit root URL.
	Slug string
	// Owner is the repository owner as read from the gitweb.owner
	// Git config key.
	Owner string
	// Description is the repository description as read from the
	// gitweb.description Git config key.
	Description string
	// LastModified records the timestamp of the most recent
	// commit as read from info/web/last-modified within the
	// repository's Git directory.
	LastModified time.Time

	// R is the raw [github.com/go-git/git-git/v5.Repository] object
	R *git.Repository
}

func NewRepo(path string, cfg config.Config) (*Repo, error) {
	var err error
	re := new(Repo)
	re.Path, _ = strings.CutPrefix(path, cfg.RepoBasePath+"/")
	re.Slug = re.Path
	if cfg.RemoveSuffix {
		re.Slug, _ = strings.CutSuffix(re.Slug, ".git")
		re.Slug, _ = strings.CutSuffix(re.Slug, "/")
	}
	if re.R, err = git.PlainOpen(path); err != nil {
		return nil, fmt.Errorf("failed to open repo %s: %v", path, err)
	}
	repoCfg, err := re.R.Config()
	if err != nil {
		log.Printf("failed to read config for repo %s: %v", path, err)
		return nil, fmt.Errorf("failed to read config for repo %s: %v", path, err)
	}
	for _, section := range repoCfg.Raw.Sections {
		if section.Name == "gitweb" {
			re.Owner = section.Option("owner")
			re.Description = section.Option("description")
		}
	}
	lastModifiedBytes, err := os.ReadFile(filepath.Join(path, "info", "web", "last-modified"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Not an error. The file just doesn't exist.
			return re, nil
		}
		log.Printf("failed to read last-modified file for repo %s: %v", path, err)
		return nil, fmt.Errorf("failed to read last-modified file for repo %s: %v", path, err)
	}
	lastModifiedBytes = bytes.TrimSpace(lastModifiedBytes)
	re.LastModified, err = time.Parse(lastModifiedFormat, string(lastModifiedBytes))
	if err != nil {
		log.Printf("failed to parse last-modified for %s: %v", path, err)
		return nil, fmt.Errorf("failed to parse last-modified for %s: %v", path, err)
	}
	return re, nil
}

// IsRepo returns true of the provided path is the base directory of a
// Git repository as determined by the presence of an objects
// directory and a HEAD file.
func IsRepo(path string) bool {
	objInfo, err := os.Stat(filepath.Join(path, "objects"))
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(path, "HEAD"))
	if err != nil {
		return false
	}
	return objInfo.IsDir()
}

type ByLastModified []*Repo

func (b ByLastModified) Len() int { return len(b) }

func (b ByLastModified) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b ByLastModified) Less(i, j int) bool {
	return b[i].LastModified.Unix() < b[j].LastModified.Unix()
}
