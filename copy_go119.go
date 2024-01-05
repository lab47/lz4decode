//go:build !go1.20
// +build !go1.20

package lz4decode

import (
	"reflect"
	"unsafe"
)

func sliceData(d []byte) *byte {
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	return (*byte)(unsafe.Pointer(hdrp.Data))
}
