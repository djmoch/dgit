// See LICENSE file for copyright and license details

// Package data implements data types that are passed to the various
// templates.
package data

import (
	"time"
)

// IndexData is provided to the index template when executed and
// becomes dot within the template.
type IndexData struct {
	Repos []*Repo
}

// IndexRepo is a single element of [IndexData] and contains data for
// a single repository.
type Repo struct {
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
}

// TreeData is provided to the head and tree templates when executed
// and becomes dot within the template.
//
// The data in TreeData is, properly speaking, a conglomeration of
// both commit and tree data. The data is combined here for ease of
// presentation.
type TreeData struct {
	// The repository slug, i.e., the URL path up to the - path element.
	Repo string
	// The base name of the Git reference, or the commit hash.
	RefOrCommit string
	// Commit information related to the tree.
	Commit Commit
	// The Tree itself.
	Tree Tree
	// Tree README contents.
	Readme string
}

// HasReadme returns true if the tree has a file named README. When
// true, the README contents are available in TreeData.Readme.
func (t TreeData) HasReadme() bool {
	return t.Readme != ""
}

// IsEmpty returns true if the Tree is empty.
func (t TreeData) IsEmpty() bool {
	return t.Commit.Hash == ""
}

// Commit contains information related to a Git commit.
type Commit struct {
	// Hash of the commit object.
	Hash string
	// Author is the original author of the commit.
	Author string
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer string
	// Message is the commit message, contains arbitrary text.
	Message string
}

// Tree contains information related to a Git tree.
type Tree struct {
	Entries []TreeEntry
	Hash    string
}

// TreeEntry contains information related to a tree entry. For
// purposes of a TreeEntry, "file" should be understood to mean blob
// or tree.
type TreeEntry struct {
	// The file name.
	Name string
	// The file mode. See [FileMode].
	Mode FileMode
	// The file hash.
	Hash string
	// The link (href) to view the file.
	Href string
}

type FileMode uint8

const (
	Empty FileMode = iota
	Dir
	File
	Executable
	Symlink
	Submodule
)

func (f FileMode) String() string {
	switch f {
	case Dir:
		return "d---------"
	case File:
		return "-rw-r--r--"
	case Executable:
		return "-rwxr-xr-x"
	case Symlink:
		return "lrwxrwxrwx"
	default:
		return "Unknown"
	}
}

// BlobData is provided to the blob template when executed and
// becomes dot within the template.
//
// The data in TreeData is, properly speaking, a conglomeration of
// both commit and tree data. The data is combined here for ease of
// presentation.
type BlobData struct {
	// The repository slug, i.e., the URL path up to the - path element.
	Repo string
	// The base name of the Git reference, or the commit hash.
	RefOrCommit string
	// The path of the blob within the repo
	Path string
	// Commit information related to the blob.
	Commit Commit
	// The Blob itself.
	Blob Blob
}

type Blob struct {
	Hash     string
	Size     int64
	Contents string
}
