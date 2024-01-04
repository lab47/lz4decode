//go:build !go1.20

package lz4decode

import "fmt"

func UncompressBlockGoFast(src, dst, dict []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	if di := decodeBlockGo(dst, src, dict); di >= 0 {
		return di, nil
	}
	return 0, fmt.Errorf("short buffers")
}
