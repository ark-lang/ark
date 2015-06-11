package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ark-lang/ark/codegen"
	"github.com/ark-lang/ark/codegen/LLVMCodegen"
	"github.com/ark-lang/ark/codegen/arkcodegen"
	"github.com/ark-lang/ark/common"
	"github.com/ark-lang/ark/doc"
	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/parser"
	"github.com/ark-lang/ark/util"
)

const (
	VERSION = "0.0.2"
	AUTHOR  = "The Ark Authors"
)

var startTime time.Time

func main() {
	startTime = time.Now()

	var command string
	var numFiles int

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case buildCom.FullCommand():
		ccArgs := []string{}
		if *buildStatic {
			ccArgs = append(ccArgs, "-static")
		}
		build(*buildInputs, *buildOutput, *buildCodegen, ccArgs)
		numFiles = len(*buildInputs)
		command = buildCom.FullCommand()

	case docgenCom.FullCommand():
		docgen(*docgenInputs, *docgenDir)
		numFiles = len(*docgenInputs)
		command = docgenCom.FullCommand()
	}

	if command != "" {
		dur := time.Since(startTime)
		fmt.Printf("%s (%d file(s), %.2fms)\n",
			util.TEXT_GREEN+util.TEXT_BOLD+fmt.Sprintf("Finished %s", command)+util.TEXT_RESET,
			numFiles, float32(dur.Nanoseconds())/1000000)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func parseFiles(files []string) []*parser.File {
	sourcefiles := make([]*common.Sourcefile, 0)

	for _, file := range files {
		input, err := common.NewSourcefile(file)
		check(err) // TODO nice error
		sourcefiles = append(sourcefiles, input)
	}

	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, file.Filename, *verbose)
	}

	parsedFiles := make([]*parser.File, 0)
	for _, file := range sourcefiles {
		parsedFiles = append(parsedFiles, parser.Parse(file, *verbose))
	}

	return parsedFiles
}

func build(input []string, output string, cg string, ccArgs []string) {
	parsedFiles := parseFiles(input)

	if cg != "none" {
		var gen codegen.Codegen

		switch cg {
		case "ark":
			gen = &arkcodegen.Codegen{}
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName: output,
				CCArgs:     ccArgs,
			}
		default:
			panic("whoops")
		}

		gen.Generate(parsedFiles, *verbose)
	}
}

func docgen(input []string, dir string) {
	gen := &doc.Docgen{
		Input: parseFiles(input),
		Dir:   dir,
	}

	gen.Generate(*verbose)
}
