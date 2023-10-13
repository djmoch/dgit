// See LICENSE file for copyright and license details

package request

import (
	"errors"
	"net/url"
	"testing"
)

func TestParse(t *testing.T) {
	urlTable := []struct {
		url *url.URL
		req *Request
	}{
		{
			url: mustParse("/testRepo"),
			req: &Request{
				Repo:    "testRepo",
				Section: "head",
			},
		},
		{
			url: mustParse("/testSection/testRepo"),
			req: &Request{
				Repo:    "testSection/testRepo",
				Section: "head",
			},
		},
		{
			url: mustParse("/testRepo/-/tree/master"),
			req: &Request{
				Repo:        "testRepo",
				Section:     "tree",
				RefOrCommit: "master",
			},
		},
		{
			url: mustParse("/testRepo/-/tree/master/test/path"),
			req: &Request{
				Repo:        "testRepo",
				Section:     "tree",
				RefOrCommit: "master",
				Path:        "test/path",
			},
		},
		{
			url: mustParse("/testRepo/-/refs"),
			req: &Request{
				Repo:    "testRepo",
				Section: "refs",
			},
		},
		{
			url: mustParse("/testRepo/-/diff/v1.0.0..v1.1.0"),
			req: &Request{
				Repo:     "testRepo",
				Section:  "diff",
				DiffFrom: "v1.0.0",
				DiffTo:   "v1.1.0",
			},
		},
	}

	for _, entry := range urlTable {
		req, err := Parse(entry.url)
		if err != nil {
			t.Fatal("unexpected error", err)
		}
		if req.Repo != entry.req.Repo {
			t.Fatal("Repo: exp=", entry.req.Repo, ", act=", req.Repo)
		}
		if req.Section != entry.req.Section {
			t.Fatal("Repo: exp=", entry.req.Section, ", act=", req.Section)
		}
		if req.RefOrCommit != entry.req.RefOrCommit {
			t.Fatal("Repo: exp=", entry.req.RefOrCommit, ", act=", req.RefOrCommit)
		}
		if req.Path != entry.req.Path {
			t.Fatal("Repo: exp=", entry.req.Path, ", act=", req.Path)
		}
		if req.From != entry.req.From {
			t.Fatal("Repo: exp=", entry.req.From, ", act=", req.From)
		}
		if req.From != entry.req.From {
			t.Fatal("Repo: exp=", entry.req.From, ", act=", req.From)
		}
		if req.DiffFrom != entry.req.DiffFrom {
			t.Fatal("Repo: exp=", entry.req.DiffFrom, ", act=", req.DiffFrom)
		}
		if req.DiffTo != entry.req.DiffTo {
			t.Fatal("Repo: exp=", entry.req.DiffTo, ", act=", req.DiffTo)
		}
	}
}

func TestRefsWithRefOrCommit(t *testing.T) {
	_, err := Parse(mustParse("/testRepo/-/refs/bad"))
	if !errors.Is(err, ErrMalformed) {
		t.Fatal("expected malformed request")
	}
}

func TestLogWithPath(t *testing.T) {
	_, err := Parse(mustParse("/testRepo/-/log/main/bad"))
	if !errors.Is(err, ErrMalformed) {
		t.Fatal("expected malformed request")
	}
}

func mustParse(rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic("url did not parse!")
	}
	return url
}
