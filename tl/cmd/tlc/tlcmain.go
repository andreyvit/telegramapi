package main

import (
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"github.com/andreyvit/telegramapi/tl/tlc"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: tlc [-pkg <package-name>] [-o <output.go>] (<source.tl> | mtproto | telegram)...\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	var pkgName string
	var outputFile string
	flag.StringVar(&pkgName, "pkg", "", "Package name (defaults to the package in the current directory)")
	flag.StringVar(&outputFile, "o", "tlschema.go", "Output file name (defaults to tlschema.go)")
	flag.Usage = Usage
	flag.Parse()

	if pkgName == "" {
		directory := "."
		pkg, err := build.ImportDir(directory, 0)
		if err != nil {
			log.Fatalf("cannot process directory %s: %s", directory, err)
		}
		pkgName = pkg.Name
	}

	if flag.NArg() == 0 {
		Usage()
		os.Exit(1)
	}

	sch := new(tlschema.Schema)

	for _, schemaName := range flag.Args() {
		var schema string
		var options tlschema.ParseOptions
		if schemaName == "mtproto" {
			schema = knownschemas.MTProtoSchema
			options.Origin = "MTProto"
			options.Alterations = &tlschema.Alterations{
				Renamings: map[string]string{
					"message": "proto_message",
					"Message": "ProtoMessage",
				},
			}
			options.FixZeroTagsIn = map[string]bool{
				"int":     true,
				"long":    true,
				"double":  true,
				"string":  true,
				"message": true,
			}
		} else if schemaName == "telegram" {
			schema = knownschemas.TelegramSchema
			options.Origin = "Telegram"
		} else if strings.Contains(schemaName, ".") {
			fmt.Fprintf(os.Stderr, "** Reading schema from file isn't implemented yet\n")
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "** Unknown schema: %s\n", schemaName)
			os.Exit(1)
		}

		err := sch.Parse(schema, options)
		if err != nil {
			fmt.Fprintf(os.Stderr, "** Failed to parse schema %s: %v\n", schemaName, err)
			os.Exit(1)
		}
	}

	options := tlc.Options{
		PackageName: pkgName,
	}
	code := tlc.GenerateGoCode(sch, options)

	if outputFile == "-" {
		fmt.Print(code)
	} else {
		err := ioutil.WriteFile(outputFile, []byte(code), 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "** Error writing to %v: %v\n", outputFile, err)
			os.Exit(1)
		}
	}
}
