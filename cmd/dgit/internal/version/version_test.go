// See LICENSE file for copyright and license details

package version

import (
	"context"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	b := new(strings.Builder)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "w", b)
	printVersion(ctx)
	if !strings.HasPrefix(b.String(), "dgit version") {
		t.Error("Unexpected version message")
	}
}
