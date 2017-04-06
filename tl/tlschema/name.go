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

func ToGoName(s string) string {
	var buf bytes.Buffer
	buf.Grow(len(s))

	up := true
	for _, r := range s {
		if r == '_' {
			up = true
		} else if up {
			buf.WriteRune(unicode.ToUpper(r))
			up = false
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

func toBareName(s string) string {
	var buf bytes.Buffer
	buf.Grow(len(s))
	buf.WriteString(strings.ToLower(s[:1]))
	buf.WriteString(s[1:])
	return buf.String()
}
