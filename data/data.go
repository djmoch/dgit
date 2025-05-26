// See LICENSE file for copyright and license details

// Package data implements data types that are passed to the various
// templates.
package data

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"log"
	"path"
	"strconv"
	"strings"
	"time"
)

// TODO(djmoch): Migrate the below variables into config.Config

var (
	// LogPageSize is the number of log entries presented per
	// page.
	LogPageSize = 20
	// DiffContext is t he number of context lines presented on
	// either side of a diff
	DiffContext = 3
)

// IndexData is provided to the index template when executed and
// becomes dot within the template.
type IndexData struct {
	// Repos is a slice of repositories
	Repos []*Repo
}

// The Repo struct contains data for a single repository.
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

// RequestData is the base type for several of the other data types.
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
//	{{ for range .PathElems -}}
//		<a href="/{{ .Repo }}/-/tree/{{ .Revision }}{{ .Path }}">{{ .Base }}</a>
//	{{- end }}
func (r RequestData) PathElems() []PathElem {
	var (
		splitPath = strings.Split(string(r.Path), "/")
		sp        = new(strings.Builder)
		elems     = make([]PathElem, len(splitPath)-1)
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

// The PathElem struct contains information for a repository path
// element.
type PathElem struct {
	// Repo is the repository slug.
	Repo string
	// Revision is the Git revision.
	Revision string
	// Path is the path within the repository, excluding the
	// current path element and any following it.
	Path string
	// Base is the current path element.
	Base string
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
	// Tree README markdown contents.
	MarkdownReadme template.HTML
}

// HasReadme returns true if the tree has a file named README. When
// true, the README contents are available in TreeData.Readme.
func (t TreeData) HasReadme() bool {
	return t.Readme != ""
}

// HasMarkdownReadme returns true if the tree has a file named
// README.md. When true, the README contents are available in
// TreeData.MarkdownReadme.
func (t TreeData) HasMarkdownReadme() bool {
	return t.MarkdownReadme != ""
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

// HasParents returns true when c has one or more parents. Otherwise
// it returns false.
func (c Commit) HasParents() bool {
	return len(c.ParentHashes) != 0
}

// Hash is a Git hash.
type Hash string

// Short returns a short version of the Git hash.
func (h Hash) Short() string {
	return fmt.Sprint(h[:7])
}

// Tree contains information related to a Git tree.
type Tree struct {
	// Entries describes the entries in the current tree.
	Entries []TreeEntry
	// Hash is the hash of the current tree.
	Hash Hash
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
	Hash Hash
	// The link (href) to view the file.
	Href string
}

// FileMode contains the encoded type of a Git tree entry.
type FileMode uint8

// These are recognized [FileMode] values.
const (
	// Empty or uninitialized
	Empty FileMode = iota
	// A directory
	Dir
	// A regular file
	File
	// An executable file
	Executable
	// A symbolic link
	Symlink
	// A Git submodule
	Submodule
)

// String returns the string representation of f. The string
// representation is similar to the output of the "ls" command on Unix
// systems.
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
		return "s---------"
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
	// If the blob is a Markdown file, rendered content goes here
	RenderedMarkdown template.HTML
}

// Blob is information related to a Git blob.
type Blob struct {
	// The blob hash
	Hash string
	// The size of the blob
	Size int64
	// The contents of the blob
	Lines []BlobLine
}

// Blob line contains the line number and contents of a single line in
// a Git blob.
type BlobLine struct {
	Number  int
	Content string
}

// RefsData is provided to the refs template when executed and becomes
// dot within the template.
type RefsData struct {
	// The repository
	Repo Repo
	// A slice containing all branch references
	Branches []Reference
	// A slice containing all tag references
	Tags []Reference
}

// Reference contains information pertaining to a Git repository reference.
type Reference struct {
	// The name of the reference
	Name string
	// The time the reference was created or last updated
	// (whichever is most recent)
	Time time.Time
}

// LogData is provided to the log template when executed and becomes
// dot within the template.
type LogData struct {
	// The repository
	Repo Repo
	// The revision
	Revision string
	// The hash from which to begin displaying the log
	FromHash Hash
	// A slice of Git commit information
	Commits []Commit
	// The hash of the first commit for the next page
	NextPage Hash
}

// HasNext returns true of l.NextPage is not empty.
func (l LogData) HasNext() bool {
	return l.NextPage != ""
}

// CommitData is provided to the commit template when executed and
// becomes dot within the template.
type CommitData struct {
	// The repository
	Repo Repo
	// The revision
	Revision string
	// The commit
	Commit Commit
	// The commit diffstat, populated from [object.FileStats.String].
	Diffstat string
	// A slice of file patches
	FilePatches []FilePatch
}

// FilePatch represents the changes to an individual file.
type FilePatch struct {
	// True if the file is binary, otherwise false.
	IsBinary bool
	// The file name
	File string
	// A slice of chunks representing changes to the file
	Chunks []Chunk
}

var errBinaryPatch = errors.New("cannot print diff for a binary patch")

// Info converts fp to a slice of PatchInfo, ideal for display within
// an HTML table.
func (fp FilePatch) Info() ([]PatchInfo, error) {
	var (
		lineNum  = 0
		lines    = make([]int, 0)
		fullInfo = make([]PatchInfo, 0)
		info     = make([]PatchInfo, 0)

		left, right = 1, 1
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
	context:
		for _, diffLine := range lines {
			switch {
			case i < diffLine && (i+DiffContext) >= diffLine,
				i == diffLine,
				i > diffLine && (i-DiffContext) <= diffLine:
				lineInContext = true
				inDiff = true
				break context
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

// String implements the [fmt.Stringer] interface for FilePatch.
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

// Chunk represents the content and type of a file patch.
type Chunk struct {
	Content string
	Type    Operation
}

// Operation describes the type of patch operation.
type Operation int

const (
	// The chunk is equal within both files
	Equal Operation = iota
	// Content is added to the old file to make the new one
	Add
	// Content is deleted from the old file to make the new one
	Delete
)

// PatchInfo represents a single line of a file patch, structured for
// display within an HTML table.
type PatchInfo struct {
	// The line numbers in the left (old) and right (new) files
	Left, Right string
	// The operation being performed in the current line
	Operation Operation
	// The content of the current line
	Content string
}

// DiffData is provided to the diff template when executed and becomes
// dot within the template.
type DiffData struct {
	// The repository
	Repo Repo
	// The source (from) and destination (to) revision
	From, To string
	// The diffstat
	Diffstat string
	// File patches
	FilePatches []FilePatch
}
