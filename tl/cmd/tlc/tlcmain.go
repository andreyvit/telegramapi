package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"github.com/andreyvit/telegramapi/tl/tlc"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

func main() {
	flag.Parse()

	if flag.NArg() < 3 {
		fmt.Fprintf(os.Stderr, "** Usage: tlc (<source.tl> | mtproto | telegram) <package-name> <output.go>\n")
		os.Exit(1)
	}

	schemaName := flag.Arg(0)
	pkgName := flag.Arg(1)
	outputFile := flag.Arg(2)

	var schema string
	if schemaName == "mtproto" {
		schema = knownschemas.MTProtoSchema
	} else if schemaName == "telegram" {
		schema = knownschemas.TelegramSchema
	} else if strings.Contains(schemaName, ".") {
		fmt.Fprintf(os.Stderr, "** Reading schema from file isn't implemented yet\n")
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "** Unknown schema: %s\n", schemaName)
		os.Exit(1)
	}

	options := tlc.Options{
		PackageName: pkgName,
	}

	sch := tlschema.MustParse(schema)

	code := tlc.GenerateGoCode(sch, options)

	err := ioutil.WriteFile(outputFile, []byte(code), 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "** Error writing to %v: %v\n", outputFile, err)
		os.Exit(1)
	}
}
