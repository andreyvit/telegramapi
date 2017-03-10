// Infinite Garble Extension (IGE) mode.

// IGE is similar to cipher block chaining (CBC) mode:
//
// 1) like in CBC, the plaintext is XORed, before encryption,
// with the previous ciphertext;
//
// 2) unlike in CBC, after going through the block cipher, the result is also
// XORed with the previous plaintext.
//
// For the first block, the corresponding inputs are taken from the IV,
// which must be twice the block size.
//
package cryptige

import (
	"crypto/cipher"
)

type ige struct {
	b         cipher.Block
	blockSize int
	iv        []byte
}

func newIGE(b cipher.Block, iv []byte) *ige {
	ivdup := make([]byte, len(iv))
	copy(ivdup, iv)

	return &ige{
		b:         b,
		blockSize: b.BlockSize(),
		iv:        ivdup,
	}
}

type igeEncrypter ige

// NewIGEEncrypter returns a BlockMode which encrypts in cipher block chaining
// mode, using the given Block. The length of iv must be the same as the
// Block's block size.
func NewIGEEncrypter(b cipher.Block, iv []byte) cipher.BlockMode {
	if len(iv) != 2*b.BlockSize() {
		panic("cipher.NewIGEDecrypter: IV length must equal twice the block size")
	}
	return (*igeEncrypter)(newIGE(b, iv))
}

func (x *igeEncrypter) BlockSize() int { return x.blockSize }

func (x *igeEncrypter) CryptBlocks(dst, src []byte) {
	bs := x.blockSize
	if len(src)%bs != 0 {
		panic("cryptoige: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("cryptoige: output smaller than input")
	}

	ct := x.iv[:bs] // previous ciphertext
	pt := x.iv[bs:] // previous plaintext

	for len(src) > 0 {
		// XOR with previous ciphertext (from src into dst)
		xorBytes(dst[:bs], src[:bs], ct)
		// block cipher (in place)
		x.b.Encrypt(dst[:bs], dst[:bs])
		// XOR with previous plaintext (in place)
		xorBytes(dst[:bs], dst[:bs], pt)

		// Move to the next block with this block as the next ct/pt.
		pt = src[:bs]
		ct = dst[:bs]
		src = src[bs:]
		dst = dst[bs:]
	}

	// Save the iv for the next CryptBlocks call.
	copy(x.iv[:bs], ct)
	copy(x.iv[bs:], pt)
}

type igeDecrypter ige

// NewIGEDecrypter returns a BlockMode which decrypts in cipher block chaining
// mode, using the given Block. The length of iv must be the same as the
// Block's block size and must match the iv used to encrypt the data.
func NewIGEDecrypter(b cipher.Block, iv []byte) cipher.BlockMode {
	if len(iv) != 2*b.BlockSize() {
		panic("cipher.NewIGEDecrypter: IV length must equal twice the block size")
	}
	return (*igeDecrypter)(newIGE(b, iv))
}

func (x *igeDecrypter) BlockSize() int { return x.blockSize }

func (x *igeDecrypter) CryptBlocks(dst, src []byte) {
	bs := x.blockSize
	if len(src)%bs != 0 {
		panic("cryptoige: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("cryptoige: output smaller than input")
	}
	if len(src) == 0 {
		return
	}

	ct := x.iv[:bs] // previous ciphertext
	pt := x.iv[bs:] // previous plaintext

	for len(src) > 0 {
		// XOR with previous plaintext (from src into dst)
		xorBytes(dst[:bs], src[:bs], pt)
		// block cipher (in place)
		x.b.Decrypt(dst[:bs], dst[:bs])
		// XOR with previous ciphertext (in place)
		xorBytes(dst[:bs], dst[:bs], ct)

		// Move to the next block with this block as the next ct/pt.
		pt = dst[:bs]
		ct = src[:bs]
		src = src[bs:]
		dst = dst[bs:]
	}

	// Save the iv for the next CryptBlocks call.
	copy(x.iv[:bs], ct)
	copy(x.iv[bs:], pt)
}

func xorBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) != n {
		panic("len mismatch in xorBytes")
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}
