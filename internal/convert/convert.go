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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrTreeNotFound = errors.New("tree not found")
	ErrBlobNotFound = errors.New("blob not found")
)

func ToIndexData(repos []*repo.Repo) data.IndexData {
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

func ToTreeData(repo *repo.Repo, req *request.Request) (data.TreeData, error) {
	var (
		t = data.TreeData{
			RequestData: data.RequestData{
				Repo:        repo.Slug,
				Path:        req.Path,
				RefOrCommit: req.RefOrCommit,
			},
		}
		readmes = make(map[string]plumbing.Hash)
	)
	hash, err := refOrCommitToHash(req.RefOrCommit, repo.R)
	if err != nil {
		return t, err
	}
	c, err := repo.R.CommitObject(hash)
	if err != nil {
		return t, fmt.Errorf("error resolving commit: %w", err)
	}
	t.Commit.Hash = data.Hash(c.Hash.String())
	t.Commit.Author = c.Author.Name
	t.Commit.Committer = c.Committer.Name
	t.Commit.Message = c.Message
	gitTree, err := repo.R.TreeObject(c.TreeHash)
	if err != nil {
		return t, fmt.Errorf("error resolving tree: %w", err)
	}
	p := req.Path
	for p != "" && p != "/" {
		gitTree, p, err = nextTree(gitTree, p, repo.R)
		if err != nil {
			return t, err
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
		case filemode.Regular, filemode.Deprecated:
			mode = data.File
		case filemode.Executable:
			mode = data.Executable
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
		switch entry.Name {
		case "README", "README.md", "README.rst":
			readmes[entry.Name] = entry.Hash
		}
	}

	if len(readmes) > 0 {
		var (
			hash plumbing.Hash
			tmp  plumbing.Hash
			ok   bool
		)
		// least preferred first
		tmp, ok = readmes["README.rst"]
		if ok {
			hash = tmp
		}
		tmp, ok = readmes["README.md"]
		if ok {
			hash = tmp
		}
		tmp, ok = readmes["README"]
		if ok {
			hash = tmp
		}
		rBlob, err := readBlob(hash, repo.R)
		if err != nil {
			return t, err
		}
		t.Readme = rBlob.Contents
	}
	return t, nil
}

func ToBlobData(repo *repo.Repo, req *request.Request) (data.BlobData, error) {
	b := data.BlobData{
		RequestData: data.RequestData{
			Repo:        repo.Slug,
			RefOrCommit: req.RefOrCommit,
			Path:        req.Path,
		},
	}
	hash, err := refOrCommitToHash(req.RefOrCommit, repo.R)
	if err != nil {
		return b, err
	}
	c, err := repo.R.CommitObject(hash)
	if err != nil {
		return b, fmt.Errorf("error resolving commit: %w", err)
	}
	b.Commit.Hash = data.Hash(c.Hash.String())
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
		gitTree, p, err = nextTree(gitTree, p, repo.R)
		if err != nil {
			return b, err
		}
	}

	baseName := path.Base(req.Path)
	for _, entry := range gitTree.Entries {
		if entry.Name == baseName {
			b.Blob, err = readBlob(entry.Hash, repo.R)
			if err != nil {
				return b, err
			}
			break
		}
	}
	return b, nil
}

func ToRefsData(repo *repo.Repo) (data.RefsData, error) {
	r := data.RefsData{
		Repo: repo.Slug,
		Tags: make([]data.Reference, 0, 0),
	}
	// TODO(dmoch): repo.R.References() might be cleaner
	bIter, err := repo.R.Branches()
	if err != nil {
		return r, fmt.Errorf("error listing branches: %w", err)
	}
	defer bIter.Close()
	if err := bIter.ForEach(func(ref *plumbing.Reference) error {
		if object, err := repo.R.CommitObject(ref.Hash()); err == nil {
			r.Branches = append(r.Branches, data.Reference{
				Name: path.Base(string(ref.Name())),
				Time: object.Committer.When,
			})
			return nil
		}
		return fmt.Errorf("error resolving branch %s: %w", ref, err)
	}); err != nil {
		return r, fmt.Errorf("error enumerating branches: %w", err)
	}

	tIter, err := repo.R.Tags()
	if err != nil {
		return r, fmt.Errorf("error listing tags: %w", err)
	}
	defer tIter.Close()
	if err := tIter.ForEach(func(ref *plumbing.Reference) error {
		if object, err := repo.R.TagObject(ref.Hash()); err == nil {
			r.Tags = append(r.Tags, data.Reference{
				Name: path.Base(string(ref.Name())),
				Time: object.Tagger.When,
			})
			return nil
		}
		if object, err := repo.R.CommitObject(ref.Hash()); err == nil {
			r.Tags = append(r.Tags, data.Reference{
				Name: path.Base(string(ref.Name())),
				Time: object.Committer.When,
			})
			return nil
		}
		return fmt.Errorf("error resolving tag %s: %w", ref, err)
	}); err != nil {
		return r, fmt.Errorf("error enumerating tags: %w", err)
	}

	return r, nil
}

func ToLogData(repo *repo.Repo, req *request.Request) (data.LogData, error) {
	l := data.LogData{
		Repo:        repo.Slug,
		RefOrCommit: req.RefOrCommit,
		Commits:     make([]data.Commit, 0, data.LogPageSize),
	}
	l.FromHash = req.From
	if req.From == "" {
		hash, err := refOrCommitToHash(req.RefOrCommit, repo.R)
		if err != nil {
			return l, err
		}
		l.FromHash = data.Hash(hash.String())
	}
	lo := &git.LogOptions{
		From:  plumbing.NewHash(string(l.FromHash)),
		Order: git.LogOrderCommitterTime,
	}
	gl, err := repo.R.Log(lo)
	defer gl.Close()
	if err != nil {
		return l, fmt.Errorf("error getting log: %w", err)
	}
	for i := 0; i < data.LogPageSize; i += 1 {
		c, err := gl.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return l, fmt.Errorf("error getting commit from log: %w", err)
		}
		commit := data.Commit{
			Hash:      data.Hash(c.Hash.String()),
			Author:    c.Author.Name,
			Committer: c.Committer.Name,
			Message:   strings.Split(c.Message, "\n")[0],
			Time:      c.Committer.When,
		}
		commit.ParentHashes = make([]data.Hash, len(c.ParentHashes),
			len(c.ParentHashes))
		for i, ph := range c.ParentHashes {
			commit.ParentHashes[i] = data.Hash(ph.String())
		}
		l.Commits = append(l.Commits, commit)
	}
	if len(l.Commits) == data.LogPageSize && l.Commits[data.LogPageSize-1].HasParents() {
		l.NextPage = l.Commits[data.LogPageSize-1].ParentHashes[0]
	}
	return l, nil
}

func nextTree(tree *object.Tree, path string, repo *git.Repository) (*object.Tree, string, error) {
	var (
		found = false
		p     = path

		err error
	)
	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Dir && strings.HasPrefix(path, entry.Name) {
			found = true
			p, _ = strings.CutPrefix(p, entry.Name)
			p, _ = strings.CutPrefix(p, "/")
			tree, err = repo.TreeObject(entry.Hash)
			if err != nil {
				return nil, "", fmt.Errorf("error resolving tree: %w", err)
			}
		}
	}
	if !found {
		return nil, "", fmt.Errorf("error locating tree %s: %w", path, ErrTreeNotFound)
	}
	return tree, p, nil
}

func readBlob(hash plumbing.Hash, repo *git.Repository) (data.Blob, error) {
	var blob data.Blob
	b, err := repo.BlobObject(hash)
	if err != nil {
		return blob, fmt.Errorf("error resolving blob %s: %w", hash, err)
	}
	breader, err := b.Reader()
	if err != nil {
		return blob, fmt.Errorf("error opening blob %s: %w", hash, err)
	}
	defer breader.Close()
	bytes, err := io.ReadAll(breader)
	if err != nil {
		return blob, fmt.Errorf("error reading blob %s: %w", hash, err)
	}
	blob.Contents = fmt.Sprintf("%s", bytes)
	blob.Hash = b.Hash.String()
	blob.Size = b.Size
	return blob, nil
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

	tObject, err := repo.TagObject(tag.Hash())
	if err != nil {
		return tag.Hash(), nil
	}
	return tObject.Target, nil
}
