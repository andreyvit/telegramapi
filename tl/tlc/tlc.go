package tlc

import (
	"bytes"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

func GenerateGoCode(sch *tlschema.Schema) string {
	buf := new(bytes.Buffer)

	for _, typ := range sch.Types() {
		buf.WriteString("type ")
		buf.WriteString(typ.Name.GoName())
		buf.WriteString(" struct {\n")
		buf.WriteString("}\n")
	}

	return buf.String()
}
