package tools

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"unsafe"
)

// copy golang.org/x/exp/constraints/constraints.go

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
type Integer interface {
	Signed | Unsigned
}
type Float interface {
	~float32 | ~float64
}
type Complex interface {
	~complex64 | ~complex128
}
type Ordered interface {
	Integer | Float | ~string
}

func Min[T Ordered](min T, arg ...T) T {
	for _, v := range arg {
		if min > v {
			min = v
		}
	}
	return min
}

func Max[T Ordered](max T, arg ...T) T {
	for _, v := range arg {
		if max < v {
			max = v
		}
	}
	return max
}

func Md5sum(s string, isFile bool) (string, error) {
	h := md5.New()
	if isFile {
		fr, err := os.Open(s)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(h, fr)
		if ec := fr.Close(); err == nil {
			err = ec
		}

		if err != nil {
			return "", err
		}
	} else {
		if _, err := h.Write([]byte(s)); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func StringToBytes(s string) []byte {
	// copy os.File#WriteString
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
	// unsafe.String(&b[0], len(b))
	// copy strings.Builder#String
	return unsafe.String(unsafe.SliceData(b), len(b))
}
