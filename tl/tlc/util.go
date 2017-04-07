package tlc

import (
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

func IDConstName(comb *tlschema.Comb) string {
	return "Tag" + comb.CombName.GoName()
}
