// Package benchhelper provides utilities for large data set memory or cpu
// benchmark.
package benchhelper

import (
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/openacid/tablewriter"
)

// Allocated returns the in-use heap in bytes.
func Allocated() int64 {
	for i := 0; i < 10; i++ {
		runtime.GC()
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return int64(stats.Alloc)
}

func NewBytesSlices(eltSize int, n int) [][]byte {
	slices := make([][]byte, n)

	for i := 0; i < n; i++ {
		slices[i] = make([]byte, eltSize)
	}

	return slices
}

func RandI32SliceBetween(min int32, max int32, factor float64) []int32 {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))

	indexes := make([]int32, 0)

	for i := min; i < max; i++ {
		if rnd.Float64() < factor {
			indexes = append(indexes, i)
		}
	}

	return indexes
}

func NewMDFileTable(fn string) (*os.File, *tablewriter.Table) {

	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	err = f.Truncate(0)
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(f)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	return f, table

}