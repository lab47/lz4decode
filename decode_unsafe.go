package lz4decode

import (
	"fmt"
	"unsafe"
)

func copy8(d []byte, di uint, s []byte) {
	*(*uint64)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*uint64)(unsafe.Pointer(sliceData(s)))
}

func copy7(d []byte, di uint, s []byte) {
	*(*[7]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[7]byte)(unsafe.Pointer(sliceData(s)))
}

func copy6(d []byte, di uint, s []byte) {
	*(*[6]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[6]byte)(unsafe.Pointer(sliceData(s)))
}

func copy5(d []byte, di uint, s []byte) {
	*(*[5]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[5]byte)(unsafe.Pointer(sliceData(s)))
}

func copy4(d []byte, di uint, s []byte) {
	*(*[4]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[4]byte)(unsafe.Pointer(sliceData(s)))
}

func copy3(d []byte, di uint, s []byte) {
	*(*[3]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[3]byte)(unsafe.Pointer(sliceData(s)))
}

func copy2(d []byte, di uint, s []byte) {
	*(*[2]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*[2]byte)(unsafe.Pointer(sliceData(s)))
}

func copy1(d []byte, di uint, s []byte) {
	d[di] = s[0]
}

func copy16(d []byte, di uint, s []byte, si uint) {
	type blah struct{ x, y uint64 }
	*(*blah)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di))) = *(*blah)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(s))) + uintptr(si)))
}

func copy18(d []byte, di uint, s []byte, si uint) {
	copy16(d, di, s, si)
	*(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(d))) + uintptr(di+16))) = *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(sliceData(s))) + uintptr(si+16)))
}

func u16S(s []byte, offset uint) uint {
	return uint(*(*uint16)(unsafe.Pointer(
		uintptr(unsafe.Pointer(sliceData(s))) + uintptr(offset),
	)))
}

func UncompressBlockGoFast(src, dst, dict []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	if di := decodeBlockGoInline2(dst, src, dict); di >= 0 {
		return di, nil
	}
	return 0, fmt.Errorf("short buffers")
}

func decodeBlockGoInline2(dst, src, dict []byte) (ret int) {
	// Restrict capacities so we don't read or write out of bounds.
	dst = dst[:len(dst):len(dst)]
	src = src[:len(src):len(src)]

	const hasError = -2

	if len(src) == 0 {
		return hasError
	}

	defer func() {
		if recover() != nil {
			ret = hasError
		}
	}()

	var si, di uint
	for si < uint(len(src)) {
		// Literals and match lengths (token).
		b := uint(src[si])
		si++

		// Literals.
		edge := si + 16
		lLen := b >> 4
		if lLen > 0 {
			switch {
			case lLen < 0xF && edge < uint(len(src)):
				// Shortcut 1
				// if we have enough room in src and dst, and the literals length
				// is small enough (0..14) then copy all 16 bytes, even if not all
				// are part of the literals.
				copy16(dst, di, src, si)
				si += lLen
				di += lLen
				if mLen := b & 0xF; mLen < 0xF {
					// Shortcut 2
					// if the match length (4..18) fits within the literals, then copy
					// all 18 bytes, even if not all are part of the literals.
					mLen += minMatch
					offset := u16S(src, si)
					i := di - offset
					if mLen <= offset && offset < di {
						// The remaining buffer may not hold 18 bytes.
						// See https://github.com/pierrec/lz4/issues/51.
						if end := i + 18; end <= uint(len(dst)) {
							copy18(dst, di, dst, i)
							si += 2
							di += mLen
							continue
						}
					}
				}
			case lLen == 0xF:
				// doubles are most common, let's handle them special
				x := uint(src[si])
				si++
				if lLen += x; int(lLen) < 0 {
					return hasError
				}

				if x != 0xff {
					// Shortcut 3a
					// Perform the copy (it's probably large because it took 2
					// bytes), then perform the short match logic like we do above.
					copy(dst[di:di+lLen], src[si:si+lLen])
					si += lLen
					di += lLen

					// Shortcut 3b
					// if the match length (4..18) fits within the literals, then copy
					// all 18 bytes, even if not all are part of the literals.
					if mLen := b & 0xF; mLen < 0xF {
						mLen += minMatch
						offset := u16S(src, si)
						i := di - offset
						if mLen <= offset && offset < di {
							// The remaining buffer may not hold 18 bytes.
							// See https://github.com/pierrec/lz4/issues/51.
							if end := i + 18; end <= uint(len(dst)) {
								copy18(dst, di, dst, i)
								si += 2
								di += mLen
								continue
							}
						}
					}
				} else {
					// Only used for 3+ length bytes
					for {
						x := uint(src[si])
						si++
						if lLen += x; int(lLen) < 0 {
							return hasError
						}
						if x != 0xFF {
							break
						}
					}

					copy(dst[di:di+lLen], src[si:si+lLen])
					si += lLen
					di += lLen
				}

			default:
				// Only used as a fallback in the case that
				// it's a short len but there isn't enough for a 16 byte copy.
				copy(dst[di:di+lLen], src[si:si+lLen])
				si += lLen
				di += lLen
			}
		}

		mLen := b & 0xF
		if si == uint(len(src)) && mLen == 0 {
			break
		} else if si >= uint(len(src)) {
			return hasError
		}

		offset := u16S(src, si)
		if offset == 0 {
			return hasError
		}
		si += 2

		// Match.
		mLen += minMatch
		if mLen == minMatch+0xF {
			for {
				x := uint(src[si])
				if mLen += x; int(mLen) < 0 {
					return hasError
				}
				si++
				if x != 0xFF {
					break
				}
			}
		}

		// Copy the match.
		if di < offset {
			// The match is beyond our block, meaning the first part
			// is in the dictionary.
			fromDict := dict[uint(len(dict))+di-offset:]
			n := uint(copy(dst[di:di+mLen], fromDict))
			di += n
			if mLen -= n; mLen == 0 {
				continue
			}
			// We copied n = offset-di bytes from the dictionary,
			// then set di = di+n = offset, so the following code
			// copies from dst[di-offset:] = dst[0:].
		}

		expanded := dst[di-offset:]
		if mLen > offset {
			// Efficiently copy the match dst[di-offset:di] into the dst slice.
			bytesToCopy := offset * (mLen / offset)
			end := bytesToCopy + offset
			if len(expanded) < int(bytesToCopy) {
				return hasError
			}

			for n := offset; n <= end; n *= 2 {
				switch n {
				case 1:
					copy1(expanded, n, expanded)
				case 2:
					copy2(expanded, n, expanded)
				case 4:
					copy4(expanded, n, expanded)
				case 8:
					copy8(expanded, n, expanded)
				default:
					copy(expanded[n:], expanded[:n])
				}
			}
			di += bytesToCopy
			mLen -= bytesToCopy
		}

		if len(dst) < int(di+mLen) {
			return hasError
		}

		switch mLen {
		case 1:
			copy1(dst, di, expanded)
		case 2:
			copy2(dst, di, expanded)
		case 3:
			copy3(dst, di, expanded)
		case 4:
			copy4(dst, di, expanded)
		case 5:
			copy5(dst, di, expanded)
		case 6:
			copy6(dst, di, expanded)
		case 7:
			copy7(dst, di, expanded)
		case 8:
			copy8(dst, di, expanded)
		default:
			copy(dst[di:di+mLen], expanded[:mLen])
		}

		di += mLen
	}

	return int(di)
}
