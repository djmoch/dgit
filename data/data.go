// See LICENSE file for copyright and license details

// Package data implements data types that are passed to the various
// templates.
package data

import (
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// IndexData is provided to the index template when executed and
// becomes dot within the template.
type IndexData struct {
	Repos []*Repo
}

// Time is [time.Time] with additional methods for human-readability
type Time time.Time

// Humanize returns a string in a human-readable, relative format,
// e.g., "3 days ago," or, "10 minutes ago".
func (t Time) Humanize() string {
	return humanize.Time(time.Time(t))
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
	LastModified Time
}

type RequestData struct {
	// The repository slug, i.e., the URL path up to the - path element.
	Repo string
	// The base name of the Git reference, or the commit hash.
	RefOrCommit string
	// The path of the tree within the repository.
	Path string
}

// PathElems returns a slice of [PathElem] objects for each directory
// in [RequestData.Path]. This is useful to build a breadcrumb, which
// might look like:
//
//	{{ for range .PathElems }}<a href="/{{ .Repo }}/-/tree/{{ .RefOrCommit }}{{ .Path }}">{{ .Base }}</a>/{{ end }}
func (r RequestData) PathElems() []PathElem {
	var (
		splitPath = strings.Split(string(r.Path), "/")
		sp        = new(strings.Builder)
		elems     = make([]PathElem, len(splitPath)-1, len(splitPath)-1)
	)
	for i := 0; i < len(splitPath)-1; i += 1 {
		sp.WriteString("/" + splitPath[i])
		elem := PathElem{
			Repo:        r.Repo,
			RefOrCommit: r.RefOrCommit,
			Path:        sp.String(),
			Base:        splitPath[i],
		}
		elems[i] = elem
	}
	return elems
}

// PathBase returns the base name of the request path. Returned data
// has the property requestData.RequestBase() =
// path.Base(requestData.Path).
func (r RequestData) PathBase() string {
	return path.Base(r.Path)
}

type PathElem struct {
	Repo        string
	RefOrCommit string
	Path        string
	Base        string
}

// TreeData extends [RequestData] and is provided to the head and tree
// templates when executed and becomes dot within the template.
//
// The data in TreeData is, properly speaking, a conglomeration of
// both commit and tree data. The data is combined here for ease of
// presentation.
type TreeData struct {
	RequestData
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

// Path is a URL path.
type Path string

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
	case Submodule:
		return "submodule-"
	default:
		return "-unknown--"
	}
}

// BlobData extends [RequestData] and is provided to the blob template when
// executed and becomes dot within the template.
//
// The data in TreeData is, properly speaking, a conglomeration of
// both commit and tree data. The data is combined here for ease of
// presentation.
type BlobData struct {
	RequestData
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
