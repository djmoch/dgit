// See LICENSE file for copyright and license details

package convert

import (
	"time"

	"djmo.ch/dgit/data"
)

type ByAge []data.Reference

func (b ByAge) Len() int { return len(b) }

func (b ByAge) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b ByAge) Less(i, j int) bool {
	return time.Time(b[i].Time).Unix() < time.Time(b[j].Time).Unix()
}
