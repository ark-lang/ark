package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ark-lang/ark-go/common"
	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/parser"
	//"github.com/ark-lang/ark-go/codegen"
	//"github.com/ark-lang/ark-go/codegen/LLVMCodegen"
)

func main() {
	verbose := true

	sourcefiles := make([]*common.Sourcefile, 0)

	arguments := os.Args[1:]
	for _, arg := range arguments {
		if strings.HasSuffix(arg, ".ark") {
			input, err := common.NewSourcefile(arg)
			check(err)
			sourcefiles = append(sourcefiles, input)
		} else {
			fmt.Println("unknown command")
		}
	}

	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, file.Filename, verbose)
	}

	parsedFiles := make([]*parser.File, 0)
	for _, file := range sourcefiles {
		parsedFiles = append(parsedFiles, parser.Parse(file.Tokens, verbose))
	}

	//gen := &LLVMCodegen.LLVMCodegen {}
	//gen.Generate()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func version() {
	fmt.Println("ark-go 2015 - experimental")
}
