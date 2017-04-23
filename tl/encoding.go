package tl

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
)

func DecodeObject(r *Reader) *Reader {
	if r.PeekCmd() == TagGzipPacked {
		r.ReadCmd()
		raw := r.ReadBlob()
		if err := r.Err(); err != nil {
			r.Fail(err)
			return nil
		}

		log.Printf("Gzipped data found: %x", raw)
		data, err := gunzip(raw)
		if err != nil {
			r.Fail(err)
			return nil
		}

		return NewReader(data)
	} else {
		return r
	}
}

func gunzip(compressed []byte) ([]byte, error) {
	decompressor, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, decompressor)
	if err != nil {
		return nil, err
	}

	err = decompressor.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
