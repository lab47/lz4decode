//go:build !go1.20
// +build !go1.20

package lz4decode

import "fmt"

// Uncompress the block using the fastest base implementation for your Go version.
func UncompressBlock(src, dst, dict []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	if di := decodeBlock(dst, src, dict); di >= 0 {
		return di, nil
	}
	return 0, fmt.Errorf("short buffers")
}
