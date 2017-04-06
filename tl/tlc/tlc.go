package tlc

import (
	"bytes"
	"fmt"
	"go/format"
	"log"

	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type Options struct {
	PackageName string
}

func GenerateGoCode(sch *tlschema.Schema, options Options) string {
	rm := NewReprMapper(sch)

	buf := new(bytes.Buffer)

	buf.WriteString("package ")
	buf.WriteString(options.PackageName)
	buf.WriteString("\n")

	var imports []string
	imports = append(imports, "github.com/andreyvit/telegramapi/tl")
	imports = append(imports, rm.GoImports()...)
	importsSet := make(map[string]bool)

	buf.WriteString("\n")
	buf.WriteString("import(\n")
	for _, s := range imports {
		if !importsSet[s] {
			importsSet[s] = true
			buf.WriteString("\t\"" + s + "\"\n")
		}
	}
	buf.WriteString(")\n")

	buf.WriteString("\n")
	buf.WriteString("const (\n")
	idx := 0
	for _, comb := range sch.Combs() {
		if comb.Tag == 0 {
			continue
		}
		buf.WriteString("\t")
		buf.WriteString(IDConstName(comb))
		if idx == 0 {
			buf.WriteString(" uint32")
		}
		buf.WriteString(" = ")
		buf.WriteString(fmt.Sprintf("0x%08x", comb.Tag))
		buf.WriteString("\n")
		idx++
	}
	buf.WriteString(")\n")

	rm.AppendGoDefs(buf)

	src := buf.Bytes()
	fmt, err := format.Source(src)
	if err != nil {
		log.Println(string(src))
		panic(err)
	}
	return string(fmt)
}

func IDConstName(comb *tlschema.Comb) string {
	return "Tag" + comb.CombName.GoName()
}
