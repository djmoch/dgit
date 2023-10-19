// See LICENSE file for copyright and license details

// Package data implements data types that are passed to the various
// templates.
package data

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	LogPageSize = 20
	DiffContext = 3
)

// IndexData is provided to the index template when executed and
// becomes dot within the template.
type IndexData struct {
	Repos []*Repo
}

// Repo is a single element of [IndexData] and contains data for a
// single repository.
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

type RequestData struct {
	// The repository
	Repo Repo
	// The base name of the Git reference, or the commit hash.
	Revision string
	// The path of the tree within the repository.
	Path string
}

// PathElems returns a slice of [PathElem] objects for each directory
// in [RequestData.Path]. This is useful to build a breadcrumb, which
// might look like:
//
//	{{ for range .PathElems }}<a href="/{{ .Repo }}/-/tree/{{ .Revision }}{{ .Path }}">{{ .Base }}</a>/{{ end }}
func (r RequestData) PathElems() []PathElem {
	var (
		splitPath = strings.Split(string(r.Path), "/")
		sp        = new(strings.Builder)
		elems     = make([]PathElem, len(splitPath)-1, len(splitPath)-1)
	)
	for i := 0; i < len(splitPath)-1; i += 1 {
		sp.WriteString("/" + splitPath[i])
		elem := PathElem{
			Repo:     r.Repo.Slug,
			Revision: r.Revision,
			Path:     sp.String(),
			Base:     splitPath[i],
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
	Repo     string
	Revision string
	Path     string
	Base     string
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
	Hash Hash
	// Author is the original author of the commit.
	Author string
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer string
	// Message is the commit message, contains arbitrary text.
	Message string
	// ParentHashes are the hash(es) of the parent commit(s)
	ParentHashes []Hash
	// Time is the commit timestamp
	Time time.Time
}

func (c Commit) HasParents() bool {
	return len(c.ParentHashes) != 0
}

// Hash is a Git hash
type Hash string

// Short returns a short version of the Git hash
func (h Hash) Short() string {
	return fmt.Sprint(h[:7])
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
	Hash  string
	Size  int64
	Lines []BlobLine
}

type BlobLine struct {
	Number  int
	Content string
}

// RefsData is provided to the refs template when executed and becomes
// dot within the template.
type RefsData struct {
	Repo     Repo
	Branches []Reference
	Tags     []Reference
}

type Reference struct {
	Name string
	Time time.Time
}

// LogData is provided to the log template when executed and becomes
// dot within the template.
type LogData struct {
	Repo     Repo
	Revision string
	FromHash Hash
	Commits  []Commit
	NextPage Hash
}

func (l LogData) HasNext() bool {
	return l.NextPage != ""
}

type CommitData struct {
	Repo        Repo
	Revision    string
	Commit      Commit
	Diffstat    string
	FilePatches []FilePatch
}

type FilePatch struct {
	IsBinary bool
	File     string
	Chunks   []Chunk
}

var errBinaryPatch = errors.New("cannot print diff for a binary patch")

func (fp FilePatch) Info() ([]PatchInfo, error) {
	var (
		lineNum  = 0
		lines    = make([]int, 0)
		fullInfo = make([]PatchInfo, 0)
		info     = make([]PatchInfo, 0)

		left, right int = 1, 1
	)
	if fp.IsBinary {
		return info, errBinaryPatch
	}

	for _, c := range fp.Chunks {
		s := bufio.NewScanner(strings.NewReader(c.Content))
		for ; s.Scan(); lineNum += 1 {
			switch c.Type {
			case Equal:
				fullInfo = append(fullInfo, PatchInfo{
					Left:      strconv.Itoa(left),
					Right:     strconv.Itoa(right),
					Operation: Equal,
					Content:   " " + s.Text(),
				})
				left += 1
				right += 1
			case Add:
				fullInfo = append(fullInfo, PatchInfo{
					Right:     strconv.Itoa(right),
					Operation: Add,
					Content:   "+" + s.Text(),
				})
				lines = append(lines, lineNum)
				right += 1
			case Delete:
				fullInfo = append(fullInfo, PatchInfo{
					Left:      strconv.Itoa(left),
					Operation: Delete,
					Content:   "-" + s.Text(),
				})
				lines = append(lines, lineNum)
				left += 1
			}
		}
		if err := s.Err(); err != nil {
			return info, fmt.Errorf("ERROR: Failure scanning FilePatch chunks: %w", err)
		}
	}

	inDiff := false
	for i, lineInfo := range fullInfo {
		var (
			lineInContext = false
		)
		for _, diffLine := range lines {
			switch {
			case i < diffLine && (i+DiffContext) >= diffLine,
				i == diffLine,
				i > diffLine && (i-DiffContext) <= diffLine:
				lineInContext = true
				inDiff = true
				break
			}
		}
		switch lineInContext {
		case true:
			info = append(info, lineInfo)
		case false:
			if inDiff {
				info = append(info, PatchInfo{Content: ". . ."})
				inDiff = false
			}
		}
	}

	if info[len(info)-1].Content == ". . ." {
		return info[:len(info)-1], nil
	}
	return info, nil
}

func (fp FilePatch) String() string {
	info, err := fp.Info()
	if err != nil {
		if errors.Is(err, errBinaryPatch) {
			return "Changes to binary file"
		}
		log.Println("ERROR: FilePatch.String:", err)
		return ""
	}

	sb := new(strings.Builder)

	for _, lineInfo := range info {
		sb.WriteString(lineInfo.Content + "\n")
	}

	return sb.String()
}

type Chunk struct {
	Content string
	Type    Operation
}

type Operation int

const (
	Equal Operation = iota
	Add
	Delete
)

type PatchInfo struct {
	Left, Right string
	Operation   Operation
	Content     string
}

type DiffData struct {
	Repo        Repo
	From, To    string
	Diffstat    string
	FilePatches []FilePatch
}
