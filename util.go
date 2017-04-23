package telegramapi

import (
	"strings"
)

func stripPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	} else {
		return ""
	}
}
