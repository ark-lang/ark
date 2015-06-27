package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ark-lang/ark/codegen"
	"github.com/ark-lang/ark/codegen/LLVMCodegen"
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

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case buildCom.FullCommand():
		ccArgs := []string{}
		if *buildStatic {
			ccArgs = append(ccArgs, "-static")
		}
		build(*buildInputs, *buildOutput, *buildCodegen, ccArgs, *buildAsm)
		printFinishedMessage(startTime, buildCom.FullCommand(), len(*buildInputs))
		if *buildRun {
			if *buildAsm {
				setupErr("Cannot use --run flag with -S flag")
			}
			run(*buildOutput)
		}

	case docgenCom.FullCommand():
		docgen(*docgenInputs, *docgenDir)
		printFinishedMessage(startTime, docgenCom.FullCommand(), len(*docgenInputs))
	}
}

func printFinishedMessage(startTime time.Time, command string, numFiles int) {
	dur := time.Since(startTime)
	fmt.Printf("%s (%d file(s), %.2fms)\n",
		util.TEXT_GREEN+util.TEXT_BOLD+fmt.Sprintf("Finished %s", command)+util.TEXT_RESET,
		numFiles, float32(dur.Nanoseconds())/1000000)
}

func setupErr(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Setup error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_SETUP)
}

func parseFiles(files []string) []*parser.Module {
	sourcefiles := make([]*lexer.Sourcefile, 0)

	for _, file := range files {
		input, err := lexer.NewSourcefile(file)
		if err != nil {
			setupErr("%s", err.Error())
		}
		sourcefiles = append(sourcefiles, input)
	}

	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, file.Filename, *verbose)
	}

	parsedFiles := make([]*parser.Module, 0)
	for _, file := range sourcefiles {
		parsedFiles = append(parsedFiles, parser.Parse(file, *verbose))
	}

	return parsedFiles
}

func build(input []string, output string, cg string, ccArgs []string, outputAsm bool) {
	parsedFiles := parseFiles(input)

	if cg != "none" {
		var gen codegen.Codegen

		switch cg {
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName: output,
				CCArgs:     ccArgs,
				OutputAsm:  outputAsm,
			}
		default:
			panic("whoops")
		}

		gen.Generate(parsedFiles, *verbose)
	}
}

func run(output string) {
	cmd := exec.Command("./" + output)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func docgen(input []string, dir string) {
	gen := &doc.Docgen{
		Input: parseFiles(input),
		Dir:   dir,
	}

	gen.Generate(*verbose)
}
