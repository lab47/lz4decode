//go:build go1.20
// +build go1.20

package lz4decode

import (
	"unsafe"
)

func sliceData(d []byte) *byte {
	return unsafe.SliceData(d)
}
