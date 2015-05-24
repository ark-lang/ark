package main

import (
	"fmt"
	"flag"
	
	"github.com/alloy-lang/alloy-go/common"
	"github.com/alloy-lang/alloy-go/lexer"
	"github.com/alloy-lang/alloy-go/parser"
)

var verboseFlag = flag.Bool("v", false, "enable verbose mode")
var inputFlag = flag.String("input", "", "input file")
var outputFlag = flag.String("output", "", "output file")

func main() {
	flag.Parse()
	
	verbose := *verboseFlag

	fmt.Println("alloy-go 2015")
	
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
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
