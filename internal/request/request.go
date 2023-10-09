// See LICENSE file for copyright and license details

package request

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"djmo.ch/dgit/data"
)

var (
	ErrMalformed      = errors.New("malformed request")
	ErrUnknownSection = errors.New("request for unknown Section")
)

type Request struct {
	Repo        string
	Section     string
	RefOrCommit string
	Path        string
	From        data.Hash
}

func Parse(url *url.URL) (*Request, error) {
	r := new(Request)
	splitPath := strings.Split(url.Path, "/")
	for len(splitPath) > 0 && splitPath[0] == "" {
		splitPath = splitPath[1:]
	}

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
		r.RefOrCommit = splitPath[1]
		fallthrough
	case 1:
		r.Section = splitPath[0]
	}

	r.From = data.Hash(url.Query().Get("from"))
	if r.From != "" && r.Section != "log" {
		return nil, fmt.Errorf("%w: 'from' in query not in 'log'", ErrMalformed)
	}
	return r, nil
}
