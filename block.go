package lz4decode

import "fmt"

const (
	// The following constants are used to setup the compression algorithm.
	minMatch   = 4  // the minimum size of the match sequence size (4 bytes)
	winSizeLog = 16 // LZ4 64Kb window size limit
	winSize    = 1 << winSizeLog
	winMask    = winSize - 1 // 64Kb window of previous data for dependent blocks

	// hashLog determines the size of the hash table used to quickly find a previous match position.
	// Its value influences the compression speed and memory usage, the lower the faster,
	// but at the expense of the compression ratio.
	// 16 seems to be the best compromise for fast compression.
	hashLog = 16
	htSize  = 1 << hashLog

	mfLimit = 10 + minMatch // The last match cannot start within the last 14 bytes.
)

// Uncompress the block using the hand coded per-platorm assembly (amd64, arm64, arm). Defaults to UncompressBlockGo if no assembly available.
func UncompressBlockAsm(src, dst, dict []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	if di := decodeBlock(dst, src, dict); di >= 0 {
		return di, nil
	}
	return 0, fmt.Errorf("short buffers")
}
