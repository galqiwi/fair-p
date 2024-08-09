package utils

import (
	"io"
	"net/http"
)

func Transfer(destination io.WriteCloser, source io.ReadCloser) int64 {
	defer destination.Close()
	defer source.Close()
	written, _ := io.Copy(destination, source)
	return written
}

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
