package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
	"log"
)

func main() {
	raw, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}

	data, err := gunzip(raw)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Decompressed: %x", data)
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

const data = `1f8b08000000000000036b114b38cd0004770b7f47fca9ff1d613ebd720f33902f7a64ab0c37903e73ec062b489e1188f90c4d2cf50c4d4df40ccd4df54c0d18763342e419a1f2ea460606865606494616566946c6295669602e10a01289307d20739950cc3533d733354431970961ae9979b2958149aa05c85c2342e682fcc08fec5e430303647399b1bbd79890b92ce8eeb544752f0b76f79ae0339709aa0fe15e33a07b8d0c90ed0511bc9686406f58e8999a815430a0d8cb8ae19f3490bda6b8ec3d0154df21ccc09002a40f5c6104b30d4a19181e4c6601d37780067201e5125e41e200a49e613190f54a85410e6a9f821f03c3847846b07a0175a0fb324a4a0a8aadf4f54bf47253f5216908a40d006fe6896f64020000`
