package mtproto

func PaddingOf(len int) int {
	rem := len % 4
	if rem == 0 {
		return 0
	} else {
		return 4 - rem
	}
}
