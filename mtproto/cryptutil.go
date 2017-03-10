package mtproto

func leftZeroPad(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	} else {
		result := make([]byte, n)
		copy(result[n-len(b):n], b)
		return result
	}
}
