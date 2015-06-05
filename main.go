package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ark-lang/ark/codegen"
	"github.com/ark-lang/ark/codegen/LLVMCodegen"
	"github.com/ark-lang/ark/codegen/arkcodegen"
	"github.com/ark-lang/ark/common"
	"github.com/ark-lang/ark/doc"
	"github.com/ark-lang/ark/lexer"
	"github.com/ark-lang/ark/parser"
	"github.com/ark-lang/ark/util"
)

func main() {
	startTime := time.Now()

	verbose := true
	codegenFlag := "llvm" // defaults to none
	docFlag := false

	sourcefiles := make([]*common.Sourcefile, 0)

	// TODO write nice arg parser, should be POSIX-based
	arguments := os.Args[1:]
	for _, arg := range arguments {
		if strings.HasSuffix(arg, ".ark") {
			input, err := common.NewSourcefile(arg)
			check(err)
			sourcefiles = append(sourcefiles, input)
		} else if strings.HasPrefix(arg, "--codegen=") {
			codegenFlag = arg[len("--codegen="):]
			switch codegenFlag {
			case "none", "llvm", "ark":
				// nothing to do
			default:
				fmt.Println("Invalid argument to --codegen:", codegenFlag)
				fmt.Println("Valid arguments: none, llvm, ark")
				os.Exit(99)
			}
		} else if arg == "--version" {
			version()
			return
		} else if arg == "-v" {
			verbose = true
		} else if arg == "--docgen" {
			docFlag = true
		} else {
			fmt.Println("Unknown command:", arg)
			os.Exit(98)
		}
	}

	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, file.Filename, verbose)
	}

	parsedFiles := make([]*parser.File, 0)
	for _, file := range sourcefiles {
		parsedFiles = append(parsedFiles, parser.Parse(file, verbose))
	}

	if docFlag {
		docgen := &doc.Docgen{
			Input: parsedFiles,
		}
		docgen.Generate(verbose)
	} else if codegenFlag != "none" {
		var gen codegen.Codegen

		switch codegenFlag {
		case "ark":
			gen = &arkcodegen.Codegen{}
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName: "out",
			}
		default:
			panic("whoops")
		}

		gen.Generate(parsedFiles, verbose)
	}

	dur := time.Since(startTime)
	fmt.Printf("%s %d file(s) (%.2fms)\n",
		util.TEXT_GREEN+util.TEXT_BOLD+"Finished compiling"+util.TEXT_RESET,
		len(sourcefiles), float32(dur.Nanoseconds())/1000000)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func version() {
	fmt.Println("ark 2015 - experimental")
}
