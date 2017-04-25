package telegramapi

import (
	"strings"
	"time"
)

func stripPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	} else {
		return ""
	}
}

func makeDate(date int) time.Time {
	if date == 0 {
		return time.Time{}
	} else {
		return time.Unix(int64(date), 0)
	}
}
