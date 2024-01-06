# lz4decode

This package contains multiple tuned implementations of the lz4 decode algorithm.

Changes in Go 1.20 improved the compiler combined with issues with sparsely
compressed data led to the authorship of the package as a derivative of
https://github.com/pierrec/lz4.

## Issues with https://github.com/pierrec/lz4

The hand-coded assembly for amd64, arm64, and arm were coded explicitly for densely
compressed data where the literals and matches were quite small (usually under 16
bytes). These techniques are extremely slow when handling sparsly compressed data
where the literals and matches can be multiple kilobytes. The Go version included
with https://github.com/pierrec/lz4 does not suffer from these issues with sparsely
compressed data, but out of the box is 2x slower than the assmebly versions.

After going through the profiling of benchmarks, it was shown that the result of
the 2x slowdown was the use of `copy` for all data copies, even single bytes.
Thusly UncompressBlockInlineCopy was born that contained specialized versions of
copy for 1 to 8 bytes using `unsafe`. This closed the gap on the assembly and
retained the huge advantage in sparsely compressed data.

## Advice

The best advice is to use 1.20 or later. The compiler changes vastly improved
the generated code such that the Go version is faster than the hand coded
assembly in all cases.

## Default

The `UncompressBlock` is the default that is likely to be the best for the current
Go version. For 1.20 and later, it's `UncompressBlockInlineCopy`, for pre 1.20, it's
`UncompressBlockAsm`.

## Per Versions

### Pre 1.20 - [Benchmark](https://github.com/lab47/lz4decode/actions/runs/7426881560/job/20211487522)

* UncompressBlockAsm: fastest for densely compressed data (ie, words list)
* UncompressBlockGo: fastest for sparsly compressed data

as a result:

* UncompressBlock == UncompressBlockAsm

### Post 1.20 - [Benchmark](https://github.com/lab47/lz4decode/actions/runs/7426881560/job/20211487889)

```
goos: linux
goarch: arm64
pkg: github.com/lab47/lz4decode
         │    asm        │               ic                    │
         │    sec/op     │   sec/op     vs base                │
Copy       38391.5n ± 0%   475.2n ± 1%  -98.76% (p=0.000 n=10)
Speckled     1.786µ ± 0%   1.018µ ± 1%  -43.03% (p=0.000 n=10)
Words        7.702µ ± 0%   6.734µ ± 0%  -12.57% (p=0.000 n=10)
geomean      8.083µ        1.482µ       -81.66%
```

UncompressBlockInlineCopy is equal to UncompressBlockAsm for densly compressed data
and up to 10x faster in sparsly compressed data. This is due to the hand-coded
assembly using 8 byte copy loops for all data rather than using 
`copy()/runtime.memmove` which are fastly faster for large blocks of bytes.
