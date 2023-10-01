// See LICENSE file for copyright and license details

// Package convert contains functions supporting the conversion of
// repository data into types used by the templates.
package convert

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"djmo.ch/dgit/data"
	"djmo.ch/dgit/internal/repo"
	"djmo.ch/dgit/internal/request"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrTreeNotFound = errors.New("tree not found")
	ErrBlobNotFound = errors.New("blob not found")
)

func ReposToIndexData(repos []*repo.Repo) data.IndexData {
	d := data.IndexData{Repos: make([]*data.Repo, len(repos), len(repos))}
	for i, repo := range repos {
		ir := &data.Repo{
			Slug:         repo.Slug,
			Owner:        repo.Owner,
			Description:  repo.Description,
			LastModified: repo.LastModified,
		}
		d.Repos[i] = ir
	}
	return d
}

func RepoToTreeData(repo *repo.Repo, req *request.Request) (data.TreeData, error) {
	t := data.TreeData{
		Repo:        repo.Slug,
		RefOrCommit: req.RefOrCommit,
	}
	hash, err := refOrCommitToHash(req.RefOrCommit, repo.R)
	if err != nil {
		return t, err
	}
	c, err := repo.R.CommitObject(hash)
	if err != nil {
		return t, fmt.Errorf("error resolving commit: %w", err)
	}
	t.Commit.Hash = c.Hash.String()
	t.Commit.Author = c.Author.Name
	t.Commit.Committer = c.Committer.Name
	t.Commit.Message = c.Message
	gitTree, err := repo.R.TreeObject(c.TreeHash)
	if err != nil {
		return t, fmt.Errorf("error resolving tree: %w", err)
	}
	p := req.Path
	for p != "" && p != "/" {
		found := false
		for _, entry := range gitTree.Entries {
			if entry.Mode == filemode.Dir && strings.HasPrefix(p, entry.Name) {
				found = true
				p, _ = strings.CutPrefix(p, entry.Name)
				p, _ = strings.CutPrefix(p, "/")
				gitTree, err = repo.R.TreeObject(entry.Hash)
				if err != nil {
					return t, fmt.Errorf("error resolving tree: %w", err)
				}
			}
		}
		if !found {
			return t, fmt.Errorf("error locating tree %s: %w", req.Path, ErrTreeNotFound)
		}
	}
	t.Tree.Hash = c.TreeHash.String()
	t.Tree.Entries = make([]data.TreeEntry, len(gitTree.Entries), len(gitTree.Entries))
	for i, entry := range gitTree.Entries {
		var (
			hrefSection = "blob"

			mode data.FileMode
		)
		switch entry.Mode {
		case filemode.Regular, filemode.Deprecated, filemode.Executable:
			mode = data.File
		case filemode.Dir:
			mode = data.Dir
		case filemode.Symlink:
			mode = data.Symlink
		case filemode.Submodule:
			mode = data.Submodule
		default:
			mode = data.Empty
		}
		if mode == data.Dir {
			hrefSection = "tree"
		}
		te := data.TreeEntry{
			Name: entry.Name,
			Mode: mode,
			Hash: entry.Hash.String(),
			Href: path.Clean(fmt.Sprintf("/%s/-/%s/%s/%s/%s", repo.Slug, hrefSection,
				t.RefOrCommit, req.Path, entry.Name)),
		}
		t.Tree.Entries[i] = te
		if entry.Name == "README" {
			b, err := repo.R.BlobObject(entry.Hash)
			if err != nil {
				return t, fmt.Errorf("error resolving README: %w", err)
			}
			breader, err := b.Reader()
			if err != nil {
				return t, fmt.Errorf("error opening README: %w", err)
			}
			defer breader.Close()
			bytes, err := io.ReadAll(breader)
			if err != nil {
				return t, fmt.Errorf("error reading README: %w", err)
			}
			t.Readme = fmt.Sprintf("%s", bytes)
		}
	}
	return t, nil
}

func RepoToBlobData(repo *repo.Repo, req *request.Request) (data.BlobData, error) {
	b := data.BlobData{
		Repo:        repo.Slug,
		RefOrCommit: req.RefOrCommit,
		Path:        req.Path,
	}
	hash, err := refOrCommitToHash(req.RefOrCommit, repo.R)
	if err != nil {
		return b, err
	}
	c, err := repo.R.CommitObject(hash)
	if err != nil {
		return b, fmt.Errorf("error resolving commit: %w", err)
	}
	b.Commit.Hash = c.Hash.String()
	b.Commit.Author = c.Author.Name
	b.Commit.Committer = c.Committer.Name
	b.Commit.Message = c.Message
	gitTree, err := repo.R.TreeObject(c.TreeHash)
	if err != nil {
		return b, fmt.Errorf("error resolving tree: %w", err)
	}
	p := req.Path
	p, _ = strings.CutSuffix(p, "/")
	for strings.Contains(p, "/") {
		found := false
		for _, entry := range gitTree.Entries {
			if entry.Mode == filemode.Dir && strings.HasPrefix(p, entry.Name) {
				found = true
				p, _ = strings.CutPrefix(p, entry.Name)
				p, _ = strings.CutPrefix(p, "/")
				gitTree, err = repo.R.TreeObject(entry.Hash)
				if err != nil {
					return b, fmt.Errorf("error resolving tree: %w", err)
				}
			}
		}
		if !found {
			return b, fmt.Errorf("error locating tree %s: %w", path.Dir(req.Path),
				ErrBlobNotFound)
		}
	}
	var (
		blob *object.Blob

		baseName = path.Base(req.Path)
	)
	found := false
	for _, entry := range gitTree.Entries {
		if entry.Name == baseName {
			found = true
			blob, err = repo.R.BlobObject(entry.Hash)
			if err != nil {
				return b, fmt.Errorf("error resolving blob: %w", err)
			}
		}
	}
	if !found {
		return b, fmt.Errorf("error locating blob %s: %w", req.Path,
			ErrBlobNotFound)
	}
	breader, err := blob.Reader()
	if err != nil {
		return b, fmt.Errorf("error opening %s: %w", req.Path, err)
	}
	defer breader.Close()
	bytes, err := io.ReadAll(breader)
	if err != nil {
		return b, fmt.Errorf("error reading %s: %w", req.Path, err)
	}
	b.Blob.Hash = blob.Hash.String()
	b.Blob.Size = blob.Size
	b.Blob.Contents = fmt.Sprintf("%s", bytes)
	return b, nil
}

func refOrCommitToHash(refOrCommit string, repo *git.Repository) (plumbing.Hash, error) {
	if plumbing.IsHash(refOrCommit) {
		return plumbing.NewHash(refOrCommit), nil
	}
	branch, err := repo.Reference(plumbing.ReferenceName(path.Join("refs", "heads", refOrCommit)), false)
	if err == nil {
		return branch.Hash(), nil
	}
	tag, err := repo.Tag(refOrCommit)
	if err != nil {
		return plumbing.NewHash(""),
			fmt.Errorf("failed to resolve ref: %s", refOrCommit)
	}
	return tag.Hash(), nil
}
