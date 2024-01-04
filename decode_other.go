//go:build (!amd64 && !arm && !arm64) || appengine || !gc || noasm
// +build !amd64,!arm,!arm64 appengine !gc noasm

package lz4decode

func decodeBlock(dst, src, dict []byte) (ret int) {
	return DecodeBlockGo(dst, src, dict)
}
