// See LICENSE file for copyright and license details

package convert

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"djmo.ch/dgit/config"
	"djmo.ch/dgit/internal/repo"
	"djmo.ch/dgit/internal/request"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type DumbCloneResponse struct {
	ContentType string
	Data        *bytes.Buffer
}

func ToCloneData(repo *repo.Repo, r *request.Request, cfg config.Config) (*DumbCloneResponse, error) {
	switch {
	case r.Path == "info/refs":
		return readRefs(repo.R)
	case r.Path == "objects/info/packs":
		packDir := filepath.Join(cfg.RepoBasePath, repo.Path, "objects/pack")
		return readPacks(packDir)
	case r.Path == "HEAD", strings.HasPrefix(r.Path, "objects/"):
		cType := "application/octet-stream"
		if r.Path == "HEAD" {
			log.Printf("client reading %s/HEAD (clone?)", repo.Slug)
			cType = "text/plain"
		}
		d, err := readFile(filepath.Join(cfg.RepoBasePath, repo.Path, r.Path))
		if err != nil {
			return nil, err
		}
		return &DumbCloneResponse{ContentType: cType, Data: d}, nil
	default:
		return nil, fmt.Errorf("unknown request path: %s", r.Path)
	}
}

func readFile(path string) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		if strings.Contains(err.Error(), "file or directory not found") {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		return nil, fmt.Errorf("unexpected error: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("ERROR: f.Close(): %v", err)
		}
	}()
	_, err = io.Copy(b, f)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return b, nil
}

func readRefs(r *git.Repository) (*DumbCloneResponse, error) {
	var (
		b = new(bytes.Buffer)
		d = &DumbCloneResponse{ContentType: "text/plain", Data: b}
	)
	rIter, err := r.References()
	if err != nil {
		return nil, fmt.Errorf("error listing references: %w", err)
	}
	defer rIter.Close()
	if err := rIter.ForEach(func(ref *plumbing.Reference) error {
		if strings.Contains(ref.Name().String(), "HEAD") {
			return nil
		}
		fmt.Fprintf(b, "%s\t%s\n", ref.Hash(), ref.Name())
		// also list annotated tag targets
		aTag, err := r.TagObject(ref.Hash())
		if err == nil {
			fmt.Fprintf(b, "%s\t%s\n", aTag.Target, "refs/tags"+aTag.Name+"^{}")
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error enumerating branches: %w", err)
	}
	return d, nil
}

func readPacks(path string) (*DumbCloneResponse, error) {
	switch s, err := os.Stat(path); {
	case err != nil:
		return nil, fmt.Errorf("%w: %s", ErrDirectoryNotFound, path)
	case !s.IsDir():
		return nil, fmt.Errorf("%s exists, but is not a directory", path)
	}
	var (
		b = new(bytes.Buffer)
		d = &DumbCloneResponse{ContentType: "text/plain", Data: b}
	)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("getrepolist: error accessing %s: %v", path, err)
		}
		if info.IsDir() {
			if info.Name() == "pack" {
				return nil
			}
			return filepath.SkipDir
		}
		if strings.HasSuffix(info.Name(), ".pack") {
			fmt.Fprintf(b, "P %s\n", info.Name())
		}
		return nil
	}
	filepath.Walk(path, walkFunc)
	b.WriteString("\n")
	return d, nil
}
