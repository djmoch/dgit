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
				Repo:     "testRepo",
				Section:  "tree",
				Revision: "master",
			},
		},
		{
			url: mustParse("/testRepo/-/tree/master/test/path"),
			req: &Request{
				Repo:     "testRepo",
				Section:  "tree",
				Revision: "master",
				Path:     "test/path",
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
		{
			url: mustParse("/testRepo/HEAD"),
			req: &Request{
				Repo:    "testRepo",
				Section: "dumbClone",
				Path:    "HEAD",
			},
		},
		{
			url: mustParse("/testRepo/info/refs"),
			req: &Request{
				Repo:    "testRepo",
				Section: "dumbClone",
				Path:    "info/refs",
			},
		},
		{
			url: mustParse("/testRepo/objects/info/idontknowwhat"),
			req: &Request{
				Repo:    "testRepo",
				Section: "dumbClone",
				Path:    "objects/info/idontknowwhat",
			},
		},
		{
			url: mustParse("/testRepo/objects/23/04082c3b4322518796a0586f3454cc803f0cfd"),
			req: &Request{
				Repo:    "testRepo",
				Section: "dumbClone",
				Path:    "objects/23/04082c3b4322518796a0586f3454cc803f0cfd",
			},
		},
		{
			url: mustParse("/testRepo/objects/pack/pack-2304082c3b4322518796a0586f3454cc803f0cfd.pack"),
			req: &Request{
				Repo:    "testRepo",
				Section: "dumbClone",
				Path:    "objects/pack/pack-2304082c3b4322518796a0586f3454cc803f0cfd.pack",
			},
		},
		{
			url: mustParse("/path/to/testRepo/objects/pack/pack-2304082c3b4322518796a0586f3454cc803f0cfd.idx"),
			req: &Request{
				Repo:    "path/to/testRepo",
				Section: "dumbClone",
				Path:    "objects/pack/pack-2304082c3b4322518796a0586f3454cc803f0cfd.idx",
			},
		},
		{
			url: mustParse("/testRepo/info/refs?service=git-upload-pack"),
			req: &Request{
				Repo:    "testRepo",
				Section: "smartClone",
				Path:    "info/refs",
			},
		},
		{
			url: mustParse("/testRepo/git-upload-pack?service=git-upload-pack"),
			req: &Request{
				Repo:    "testRepo",
				Section: "smartClone",
				Path:    "git-upload-pack",
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
			t.Fatal("Section: exp=", entry.req.Section, ", act=", req.Section)
		}
		if req.Revision != entry.req.Revision {
			t.Fatal("Revision: exp=", entry.req.Revision, ", act=", req.Revision)
		}
		if req.Path != entry.req.Path {
			t.Fatal("Path: exp=", entry.req.Path, ", act=", req.Path)
		}
		if req.From != entry.req.From {
			t.Fatal("From: exp=", entry.req.From, ", act=", req.From)
		}
		if req.From != entry.req.From {
			t.Fatal("From: exp=", entry.req.From, ", act=", req.From)
		}
		if req.DiffFrom != entry.req.DiffFrom {
			t.Fatal("DiffFrom: exp=", entry.req.DiffFrom, ", act=", req.DiffFrom)
		}
		if req.DiffTo != entry.req.DiffTo {
			t.Fatal("DiffTo: exp=", entry.req.DiffTo, ", act=", req.DiffTo)
		}
	}
}

func TestRefsWithRevision(t *testing.T) {
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
