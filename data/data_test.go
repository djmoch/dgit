// See LICENSE file for copyright and license details

package data

import (
	"fmt"
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
		t.Error(fmt.Sprintf("expected 3 PathElems, but got %d", len(elems)))
	}
	if elems[0].Repo != "testRepo" {
		t.Error(fmt.Sprintf("expected testRepo, but got %s", elems[0].Repo))
	}
	if elems[0].Revision != "main" {
		t.Error(fmt.Sprintf("expected main, but got %s", elems[0].Revision))
	}
	if elems[0].Path != "/test" {
		t.Error(fmt.Sprintf("expected /test, but got %s", elems[0].Path))
	}
	if elems[0].Base != "test" {
		t.Error(fmt.Sprintf("expected test, but got %s", elems[0].Base))
	}
	if elems[2].Path != "/test/path/to" {
		t.Error(fmt.Sprintf("expected /test, but got %s", elems[0].Path))
	}
	if elems[2].Base != "to" {
		t.Error(fmt.Sprintf("expected test, but got %s", elems[0].Base))
	}
}
