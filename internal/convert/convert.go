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
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrDirectoryNotFound = errors.New("directory not found")
	ErrFileNotFound      = errors.New("file not found")
)

func ToIndexData(repos []*repo.Repo) data.IndexData {
	d := data.IndexData{Repos: make([]*data.Repo, len(repos), len(repos))}
	for i, repo := range repos {
		ir := toDataRepo(repo)
		d.Repos[i] = &ir
	}
	return d
}

func ToTreeData(repo *repo.Repo, req *request.Request) (data.TreeData, error) {
	var (
		t = data.TreeData{
			RequestData: data.RequestData{
				Repo:     toDataRepo(repo),
				Path:     req.Path,
				Revision: req.Revision,
			},
		}
		readmes = make(map[string]plumbing.Hash)
	)
	hash, err := toCommitHash(req.Revision, repo.R)
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
		return t, fmt.Errorf("error resolving commit tree: %w", err)
	}
	if req.Path != "/" && req.Path != "" {
		gitTree, err = gitTree.Tree(req.Path)
		if err != nil {
			if errors.Is(err, object.ErrDirectoryNotFound) {
				return t, fmt.Errorf("%w: %s", ErrDirectoryNotFound, req.Path)
			}
			return t, fmt.Errorf("error resolving tree: %s", err)
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
				t.Revision, req.Path, entry.Name)),
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
			Repo:     toDataRepo(repo),
			Revision: req.Revision,
			Path:     req.Path,
		},
	}
	hash, err := toCommitHash(req.Revision, repo.R)
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
	f, err := c.File(req.Path)
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return b, fmt.Errorf("%w: %s", ErrFileNotFound, req.Path)
		}
		return b, fmt.Errorf("error resolving file: %w", err)
	}
	b.Blob.Hash = f.Hash.String()
	b.Blob.Size = f.Size
	b.Blob.Contents, err = f.Contents()
	if err != nil {
		return b, err
	}
	return b, nil
}

func ToRefsData(repo *repo.Repo) (data.RefsData, error) {
	r := data.RefsData{
		Repo: toDataRepo(repo),
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
		Repo:     toDataRepo(repo),
		Revision: req.Revision,
		Commits:  make([]data.Commit, 0, data.LogPageSize),
	}
	l.FromHash = req.From
	if req.From == "" {
		hash, err := toCommitHash(req.Revision, repo.R)
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

func ToCommitData(repo *repo.Repo, req *request.Request) (data.CommitData, error) {
	c := data.CommitData{
		Repo:     toDataRepo(repo),
		Revision: req.Revision,
	}
	hash, err := toCommitHash(req.Revision, repo.R)
	if err != nil {
		return c, err
	}
	gc, err := repo.R.CommitObject(hash)
	if err != nil {
		return c, fmt.Errorf("error resolving commit: %w", err)
	}
	c.Commit.Hash = data.Hash(gc.Hash.String())
	c.Commit.Author = gc.Author.Name
	c.Commit.Committer = gc.Committer.Name
	c.Commit.Message = gc.Message
	c.Commit.Time = gc.Committer.When
	fileStats, err := gc.Stats()
	if err != nil {
		return c, fmt.Errorf("error getting stats for commit: %w", err)
	}
	c.Diffstat = fileStats.String()
	c.Commit.ParentHashes = make([]data.Hash, len(gc.ParentHashes),
		len(gc.ParentHashes))
	for i, ph := range gc.ParentHashes {
		c.Commit.ParentHashes[i] = data.Hash(ph.String())
	}
	switch len(gc.ParentHashes) {
	case 0:
		files, err := gc.Files()
		c.FilePatches = make([]data.FilePatch, 0)
		if err != nil {
			return c, fmt.Errorf("error getting commit files: %w", err)
		}
		if err = files.ForEach(func(f *object.File) error {
			isBinary, err := f.IsBinary()
			if err != nil {
				return fmt.Errorf("IsBinary error: %w", err)
			}
			contents, err := f.Contents()
			if err != nil {
				return fmt.Errorf("Contents error: %w", err)
			}
			fp := data.FilePatch{
				IsBinary: isBinary,
				File:     f.Name + " (created)",
				Chunks: []data.Chunk{
					data.Chunk{Content: contents, Type: data.Add},
				},
			}
			c.FilePatches = append(c.FilePatches, fp)
			return nil
		}); err != nil {
			return c, fmt.Errorf("error reading commit files: %w", err)
		}
	default:
		pcHash := gc.ParentHashes[0]
		pc, err := repo.R.CommitObject(pcHash)
		if err != nil {
			return c, fmt.Errorf("error resolving parent commit: %w", err)
		}
		patch, err := pc.Patch(gc)
		if err != nil {
			return c, fmt.Errorf("error resolving commit patch: %w", err)
		}
		c.FilePatches = toFilePatches(patch.FilePatches())
	}
	return c, nil
}

func ToDiffData(repo *repo.Repo, req *request.Request) (data.DiffData, error) {
	d := data.DiffData{
		Repo: toDataRepo(repo),
		From: req.DiffFrom,
		To:   req.DiffTo,
	}
	hash, err := repo.R.ResolveRevision(plumbing.Revision(req.DiffFrom))
	if err != nil {
		return d, fmt.Errorf("error resolving 'from' commit hash: %w", err)
	}
	fromCommit, err := repo.R.CommitObject(*hash)
	if err != nil {
		return d, fmt.Errorf("error resolving 'from' commit: %w", err)
	}
	hash, err = repo.R.ResolveRevision(plumbing.Revision(req.DiffTo))
	if err != nil {
		return d, fmt.Errorf("error resolving 'to' commit hash: %w", err)
	}
	toCommit, err := repo.R.CommitObject(*hash)
	if err != nil {
		return d, fmt.Errorf("error resolving 'to' commit: %w", err)
	}
	patch, err := fromCommit.Patch(toCommit)
	if err != nil {
		return d, fmt.Errorf("error generating patch: %w", err)
	}
	d.Diffstat = patch.Stats().String()
	d.FilePatches = toFilePatches(patch.FilePatches())
	return d, nil
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

func toCommitHash(rev string, repo *git.Repository) (plumbing.Hash, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return plumbing.NewHash(""), fmt.Errorf("failed to resolve revision: %s", rev)
	}
	return *hash, nil
}

func toFilePatches(dPatches []diff.FilePatch) []data.FilePatch {
	patches := make([]data.FilePatch, len(dPatches), len(dPatches))
	for i, dp := range dPatches {
		chunks := dp.Chunks()
		p := data.FilePatch{
			IsBinary: dp.IsBinary(),
			Chunks:   make([]data.Chunk, len(chunks), len(chunks)),
		}
		from, to := dp.Files()
		switch {
		case from == nil:
			p.File = to.Path() + " (created)"
		case to == nil:
			p.File = from.Path() + " (deleted)"
		case from.Path() == to.Path():
			p.File = from.Path()
		default:
			p.File = from.Path() + " --> " + to.Path()
		}
		for j, pc := range dp.Chunks() {
			c := data.Chunk{
				Content: pc.Content(),
				Type:    data.Operation(pc.Type()),
			}
			p.Chunks[j] = c
		}
		patches[i] = p
	}
	return patches
}

func toDataRepo(repo *repo.Repo) data.Repo {
	return data.Repo{
		Slug:         repo.Slug,
		Owner:        repo.Owner,
		Description:  repo.Description,
		LastModified: repo.LastModified,
	}
}
