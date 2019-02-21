// Package array implements functions for the manipulation of compacted array.
package array

import (
	"errors"
	"reflect"

	"github.com/openacid/slim/bit"
	"github.com/openacid/slim/prototype"
)

// CompactedArray is a space efficient array implementation.
//
// Unlike a normal array, it does not allocate space for a element that there is
// not data in it.
type CompactedArray struct {
	prototype.CompactedArray
	Converter
}

var bmWidth = uint32(64) // how many bits of an uint64 == 2 ^ 6

func bmBit(idx uint32) (uint32, uint32) {
	c := idx >> uint32(6) // == idx / bmWidth
	r := idx & uint32(63) // == idx % bmWidth
	return c, r
}

func (sa *CompactedArray) appendElt(index uint32, elt []byte) {
	iBm, iBit := bmBit(index)

	var bmWord = &sa.Bitmaps[iBm]
	if *bmWord == 0 {
		sa.Offsets[iBm] = sa.Cnt
	}

	*bmWord |= uint64(1) << iBit
	sa.Elts = append(sa.Elts, elt...)

	sa.Cnt++
}

// ErrIndexLen is returned if number of indexes does not equal the number of
// datas, when initializing a CompactedArray.
var ErrIndexLen = errors.New("the length of index and elts must be equal")

// ErrIndexNotAscending means both indexes and datas for initialize a
// CompactedArray must be in ascending order.
var ErrIndexNotAscending = errors.New("index must be an ascending slice")

// Init initializes a compacted array from the slice type elts
// the index parameter must be a ascending array of type unit32,
// otherwise, return the ErrIndexNotAscending error
func (sa *CompactedArray) Init(index []uint32, _elts interface{}) error {

	rElts := reflect.ValueOf(_elts)
	if rElts.Kind() != reflect.Slice {
		panic("input is not a slice")
	}

	nElts := rElts.Len()

	if len(index) != nElts {
		return ErrIndexLen
	}

	capacity := uint32(0)
	if len(index) > 0 {
		capacity = index[len(index)-1] + 1
	}

	bmCnt := (capacity + bmWidth - 1) / bmWidth

	sa.Bitmaps = make([]uint64, bmCnt)
	sa.Offsets = make([]uint32, bmCnt)
	sa.Elts = make([]byte, 0, nElts*sa.GetMarshaledSize(nil))

	var prevIndex uint32
	for i := 0; i < len(index); i++ {
		if i > 0 && index[i] <= prevIndex {
			return ErrIndexNotAscending
		}

		ee := rElts.Index(i).Interface()
		sa.appendElt(index[i], sa.Marshal(ee))

		prevIndex = index[i]
	}

	return nil
}

// Get returns the value indexed by idx if it is in array, else return nil
func (sa *CompactedArray) Get(idx uint32) interface{} {
	iBm, iBit := bmBit(idx)

	if iBm >= uint32(len(sa.Bitmaps)) {
		return nil
	}

	var bmWord = sa.Bitmaps[iBm]

	if ((bmWord >> iBit) & 1) == 0 {
		return nil
	}

	base := sa.Offsets[iBm]
	cnt1 := bit.PopCnt64Before(bmWord, iBit)

	stIdx := uint32(sa.GetMarshaledSize(nil)) * (base + cnt1)

	_, val := sa.Unmarshal(sa.Elts[stIdx:])
	return val
}

// Has returns true if idx is in array, else return false
func (sa *CompactedArray) Has(idx uint32) bool {
	iBm, iBit := bmBit(idx)

	if iBm >= uint32(len(sa.Bitmaps)) {
		return false
	}

	var bmWord = sa.Bitmaps[iBm]

	return (bmWord>>iBit)&1 > 0
}
