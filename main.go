package main

import (
	"fmt"
	"flag"
	
	"github.com/ark-lang/ark-go/common"
	"github.com/ark-lang/ark-go/lexer"
	"github.com/ark-lang/ark-go/parser"
	//"github.com/ark-lang/ark-go/codegen"
	//"github.com/ark-lang/ark-go/codegen/LLVMCodegen"
)

var versionFlag = flag.Bool("version", false, "show version information")
var verboseFlag = flag.Bool("v", false, "enable verbose mode")
var inputFlag = flag.String("input", "", "input file")
var outputFlag = flag.String("output", "", "output file")

func main() {
	flag.Parse()
	
	if *versionFlag {
		version()
		return
	}
	
	verbose := *verboseFlag
	
	sourcefiles := make([]*common.Sourcefile, 0)
	input, err := common.NewSourcefile(*inputFlag)
	check(err)
	sourcefiles = append(sourcefiles, input)
	
	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, *inputFlag, verbose)
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
