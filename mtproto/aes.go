package mtproto

import (
	"crypto/aes"
	"crypto/sha1"
	"errors"
	"github.com/andreyvit/telegramapi/cryptige"
	"io"
)

var ErrNotBlockSizeMultiple = errors.New("input not multiple of block size")

var ErrHashMismatch = errors.New("encrypted data hash does not match")

func Pad(src []byte, blockSize int, random io.Reader) ([]byte, error) {
	n := len(src)
	if n%blockSize == 0 {
		return src, nil
	} else {
		pad := blockSize - (n % blockSize)
		padded := make([]byte, len(src)+pad)
		copy(padded, src)

		_, err := io.ReadFull(random, padded[len(src):])
		if err != nil {
			return nil, err
		}

		return padded, nil
	}
}

func AESIGEPadEncrypt(dst, src, key, iv []byte, random io.Reader) ([]byte, error) {
	ciph, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := ciph.BlockSize()
	if random != nil {
		src, err = Pad(src, bs, random)
		if err != nil {
			return nil, err
		}
	} else if len(src)%bs != 0 {
		return nil, ErrNotBlockSizeMultiple
	}

	if dst == nil {
		dst = make([]byte, len(src))
	}

	mode := cryptige.NewIGEEncrypter(ciph, iv)
	mode.CryptBlocks(dst, src)

	return dst, nil
}

func AESIGEDecrypt(dst, src, key, iv []byte) ([]byte, error) {
	ciph, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(src)%ciph.BlockSize() != 0 {
		return nil, ErrNotBlockSizeMultiple
	}

	if dst == nil {
		dst = make([]byte, len(src))
	}

	mode := cryptige.NewIGEDecrypter(ciph, iv)
	mode.CryptBlocks(dst, src)

	return dst, nil
}

func AESIGEPadEncryptWithHash(dst, src, key, iv []byte, random io.Reader) ([]byte, error) {
	h := sha1.Sum(src)
	hs := len(h)
	bs := aes.BlockSize

	pad := bs - ((hs + len(src)) % bs)
	if pad == bs {
		pad = 0
	}
	padded := make([]byte, hs+len(src)+pad)
	copy(padded[0:hs], h[:])
	copy(padded[hs:hs+len(src)], src)
	if pad > 0 {
		_, err := io.ReadFull(random, padded[hs+len(src):])
		if err != nil {
			return nil, err
		}
	}

	return AESIGEPadEncrypt(dst, padded, key, iv, nil)
}

func AESIGEDecryptWithHash(dst, src, key, iv []byte) ([]byte, []byte, error) {
	var err error
	dst, err = AESIGEDecrypt(dst, src, key, iv)
	if err != nil {
		return nil, nil, err
	}

	return dst[sha1.Size:], dst[:sha1.Size], nil
}
