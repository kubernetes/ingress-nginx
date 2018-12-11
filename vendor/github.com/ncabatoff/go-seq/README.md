# Package for ordering of Go values

[![GoDoc](https://godoc.org/github.com/ncabatoff/go-seq/seq?status.svg)][godoc]
[![Build Status](https://travis-ci.org/ncabatoff/go-seq.svg?branch=master)][travis]
[![Coverage Status](https://coveralls.io/repos/github/ncabatoff/go-seq/badge.svg?branch=master)](https://coveralls.io/github/ncabatoff/go-seq?branch=master)

This package is intended to allow ordering (most) values via reflection,
much like go-cmp allows comparing values for equality.

This is helpful when trying to write `Less()` methods for sorting structures.
Provided all the types in question are supported, you can simply define it
as follows:

```go
import "github.com/ncabatoff/go-seq/seq"

type (
    MyType struct {
        a int
        b string
    }
    MyTypeSlice []MyType
)
func (m MyTypeSlice) Less(i, j int) bool { return seq.Compare(m[i], m[j]) < 0 }
```

Noteable unsupported types include the builtin `complex` type, channels,
functions, and maps with non-string keys. Pointers can be ordered if their
underlying types are orderable.

[godoc]: https://godoc.org/github.com/ncabatoff/go-seq/seq
[travis]: https://travis-ci.org/ncabatoff/go-seq

## Install

```
go get -u github.com/ncabatoff/go-seq/seq
```

## License

MIT - See [LICENSE][license] file

[license]: https://github.com/ncabatoff/go-seq/blob/master/LICENSE
