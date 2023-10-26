// See LICENSE file for copyright and license details

package request

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"djmo.ch/dgit/data"
)

var (
	ErrMalformed      = errors.New("malformed request")
	ErrUnknownSection = errors.New("request for unknown Section")
)

const WebSections = "head tree blob raw diff refs log commit"

type Request struct {
	Repo             string
	Section          string
	Revision         string
	Path             string
	From             data.Hash
	DiffFrom, DiffTo string
}

var errInvalidClonePath = errors.New("invalid clone request path")

func Parse(url *url.URL) (*Request, error) {
	// We first attempt to parse the request as a clone request
	// because clone requests have stricter rules about allowable
	// paths after the repository name.
	r, err := parseCloneRequest(url)
	if err == nil {
		return r, nil
	}
	return parseWebRequest(url)
}

var (
	objectPath = regexp.MustCompile(`^objects/[0-9a-f]{2}/[0-9a-f]{38}$`)
	packPath   = regexp.MustCompile(`^objects/pack/pack-[0-9a-f]{40}.(pack|idx)$`)
)

func parseCloneRequest(url *url.URL) (*Request, error) {
	r := new(Request)
	splitPath := splitPath(url.Path)

	if len(splitPath) < 2 {
		return nil, fmt.Errorf("%w: %s", errInvalidClonePath, url.Path)
	}

	for len(splitPath) > 1 {
		var done bool
		r.Repo = path.Join([]string{r.Repo, splitPath[0]}...)
		splitPath = splitPath[1:]

		switch len(splitPath) {
		case 1:
			if splitPath[0] == "HEAD" {
				done = true
			}
		case 2:
			if path.Join(splitPath[:2]...) == "info/refs" {
				done = true
			}
		case 3:
			testPath := []byte(path.Join(splitPath[:3]...))
			if path.Join(splitPath[:2]...) == "objects/info" ||
				objectPath.Match(testPath) || packPath.Match(testPath) {
				done = true
			}
		}
		if done == true {
			r.Section = "dumbClone"
			r.Path = path.Join(splitPath...)
			return r, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", errInvalidClonePath, url.Path)
}

func parseWebRequest(url *url.URL) (*Request, error) {
	r := new(Request)
	splitPath := splitPath(url.Path)

	if len(splitPath) == 0 {
		r.Section = "repo"
		return r, nil
	}
	r.Repo = splitPath[0]
	splitPath = splitPath[1:]

	for len(splitPath) > 0 {
		var done bool
		if splitPath[0] != "-" {
			r.Repo = path.Join(r.Repo, splitPath[0])
		} else {
			done = true
		}
		splitPath = splitPath[1:]
		if done {
			break
		}
	}

	if len(splitPath) == 0 {
		r.Section = "head"
		return r, nil
	}

	switch len(splitPath) {
	default:
		r.Path = strings.Join(splitPath[2:], "/")
		fallthrough
	case 2:
		r.Revision = splitPath[1]
		fallthrough
	case 1:
		r.Section = splitPath[0]
	}

	if r.Section == "diff" {
		ids := strings.Split(r.Revision, "..")
		if len(ids) != 2 {
			return nil, fmt.Errorf("%w: bad commit range: %s",
				ErrMalformed, r.Revision)
		}
		r.Revision = ""
		r.DiffFrom = ids[0]
		r.DiffTo = ids[1]
		return r, nil
	}

	r.From = data.Hash(url.Query().Get("from"))
	if r.From != "" && r.Section != "log" {
		return nil, fmt.Errorf("%w: 'from' in query not in 'log'", ErrMalformed)
	}

	switch r.Section {
	case "refs":
		if r.Revision != "" {
			return nil, fmt.Errorf("%w: 'Revision' specified with '%s'",
				ErrMalformed, r.Section)
		}
		fallthrough
	case "log", "commit":
		if r.Path != "" {
			return nil, fmt.Errorf("%w: 'Revision' or 'Path' specified with '%s'",
				ErrMalformed, r.Section)
		}
	}

	return r, nil
}

func splitPath(path string) []string {
	splitPath := strings.Split(path, "/")
	for len(splitPath) > 0 && splitPath[0] == "" {
		splitPath = splitPath[1:]
	}
	return splitPath
}
