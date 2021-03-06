package tlschema

import (
	"bytes"
	"strings"
	"unicode"
)

type ScopedName struct {
	full     string
	shortidx int
}

func (n ScopedName) String() string {
	return n.full
}

func (n ScopedName) Full() string {
	return n.full
}

func (n ScopedName) Short() string {
	return n.full[n.shortidx:]
}

func (n ScopedName) Scope() string {
	if n.shortidx == 0 {
		return ""
	}
	return n.full[:n.shortidx-1]
}

func (n ScopedName) HasScope() bool {
	return n.shortidx != 0
}

func (n ScopedName) IsBare() bool {
	for _, r := range n.Short() {
		return !unicode.IsUpper(r)
	}
	panic("ScopedName.IsBare does not support empty names")
}

func (n *ScopedName) Alter(alter *Alterations) {
	if alter == nil || alter.Renamings == nil {
		return
	}
	if nn := alter.Renamings[n.full]; nn != "" {
		*n = MakeScopedName(nn)
	}
}

func MakeScopedNameComponents(scope, short string) ScopedName {
	if scope == "" {
		return ScopedName{short, 0}
	} else {
		return ScopedName{scope + "." + short, len(scope) + 1}
	}
}

func MakeScopedName(s string) ScopedName {
	i := strings.IndexRune(s, '.')
	if i < 0 {
		return ScopedName{s, 0}
	} else {
		return ScopedName{s, i + 1}
	}
}

func (n ScopedName) GoName() string {
	if n.HasScope() {
		return ToGoName(n.Scope()) + ToGoName(n.Short())
	} else {
		return ToGoName(n.Short())
	}
}

var abbrevs = [][]byte{[]byte("DH"), []byte("OK"), []byte("PQ"), []byte("RPC"), []byte("IP"), []byte("DC"), []byte("ID"), []byte("IDs"), []byte("API"), []byte("IPv6"), []byte("TCPo"), []byte("TCP"), []byte("URL")}
var badAbbrevs [][]byte

func init() {
	badAbbrevs = make([][]byte, len(abbrevs))
	for i, a := range abbrevs {
		bad := make([]byte, len(a))
		for k, b := range a {
			if k == 0 {
				bad[k] = byte(unicode.ToUpper(rune(b)))
			} else {
				bad[k] = byte(unicode.ToLower(rune(b)))
			}
		}
		badAbbrevs[i] = bad
	}
}

func checkAbbrevs(suffix []byte) {
	n := len(suffix)
	for i, a := range badAbbrevs {
		if len(a) == n && bytes.Equal(a, suffix) {
			copy(suffix, abbrevs[i])
			return
		}
	}
}

func ToGoName(s string) string {
	buf := make([]byte, 0, len(s))

	up := true
	cnt := 0
	for _, r := range s {
		if r == '_' {
			if cnt > 0 {
				checkAbbrevs(buf[len(buf)-cnt:])
			}
			up = true
		} else if up {
			if cnt > 0 {
				checkAbbrevs(buf[len(buf)-cnt:])
			}
			buf = append(buf, byte(unicode.ToUpper(r)))
			up = false
			cnt = 1
		} else {
			if unicode.IsUpper(r) {
				if cnt > 0 {
					checkAbbrevs(buf[len(buf)-cnt:])
				}
				cnt = 1
			} else {
				cnt++
			}
			buf = append(buf, byte(r))
		}
	}
	if cnt > 0 {
		checkAbbrevs(buf[len(buf)-cnt:])
	}

	return string(buf)
}

func toBareName(s string) string {
	var buf bytes.Buffer
	buf.Grow(len(s))
	buf.WriteString(strings.ToLower(s[:1]))
	buf.WriteString(s[1:])
	return buf.String()
}
