// See LICENSE file for copyright and license details

package data

import (
	"testing"
)

func TestPathElems(t *testing.T) {
	var (
		d = &RequestData{
			Repo:     Repo{Slug: "testRepo"},
			Revision: "main",
			Path:     "test/path/to/file",
		}
	)
	elems := d.PathElems()
	if len(elems) != 3 {
		t.Errorf("expected 3 PathElems, but got %d", len(elems))
	}
	if elems[0].Repo != "testRepo" {
		t.Errorf("expected testRepo, but got %s", elems[0].Repo)
	}
	if elems[0].Revision != "main" {
		t.Errorf("expected main, but got %s", elems[0].Revision)
	}
	if elems[0].Path != "/test" {
		t.Errorf("expected /test, but got %s", elems[0].Path)
	}
	if elems[0].Base != "test" {
		t.Errorf("expected test, but got %s", elems[0].Base)
	}
	if elems[2].Path != "/test/path/to" {
		t.Errorf("expected /test, but got %s", elems[0].Path)
	}
	if elems[2].Base != "to" {
		t.Errorf("expected test, but got %s", elems[0].Base)
	}
}
