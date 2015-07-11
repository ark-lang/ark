package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ark-lang/ark/src/util/log"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ark-lang/ark/src/codegen"
	"github.com/ark-lang/ark/src/codegen/LLVMCodegen"
	"github.com/ark-lang/ark/src/doc"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

const (
	VERSION = "0.0.2"
	AUTHOR  = "The Ark Authors"
)

var startTime time.Time

func main() {
	startTime = time.Now()

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	log.SetLevel(*logLevel)
	log.SetTags(*logTags)

	switch command {
	case buildCom.FullCommand():
		if len(*buildInputs) == 0 {
			setupErr("No input files passed.")
		}

		ccArgs := []string{}
		if *buildStatic {
			ccArgs = append(ccArgs, "-static")
		}

		outputType := parseOutputType(*buildOutputType)
		build(*buildInputs, *buildOutput, *buildCodegen, ccArgs, outputType)
		printFinishedMessage(startTime, buildCom.FullCommand(), len(*buildInputs))
		if *buildRun {
			if outputType != LLVMCodegen.OUTPUT_EXECUTABLE {
				setupErr("Can only use --run flag when building executable")
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
	log.Info("main", "%s (%d file(s), %.2fms)\n",
		util.TEXT_GREEN+util.TEXT_BOLD+fmt.Sprintf("Finished %s", command)+util.TEXT_RESET,
		numFiles, float32(dur.Nanoseconds())/1000000)
}

func setupErr(err string, stuff ...interface{}) {
	log.Error("main", util.TEXT_RED+util.TEXT_BOLD+"Setup error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_SETUP)
}

func parseFiles(files []string) ([]*parser.Module, map[string]*parser.Module) {
	sourcefiles := make([]*lexer.Sourcefile, 0)

	for _, file := range files {
		input, err := lexer.NewSourcefile(file)
		if err != nil {
			setupErr("%s", err.Error())
		}
		sourcefiles = append(sourcefiles, input)
	}

	for _, file := range sourcefiles {
		file.Tokens = lexer.Lex(file.Contents, file.Name)
	}

	parsedFiles := make([]*parser.Module, 0)
	modules := make(map[string]*parser.Module, 0)

	for _, file := range sourcefiles {
		parsedFiles = append(parsedFiles, parser.Parse(file, modules))
	}

	return parsedFiles, modules
}

func build(files []string, outputFile string, cg string, ccArgs []string, outputType LLVMCodegen.OutputType) {
	// read source files
	var sourcefiles []*lexer.Sourcefile

	timed("reading sourcefiles", func() {
		for _, file := range files {
			sourcefile, err := lexer.NewSourcefile(file)
			if err != nil {
				setupErr("%s", err.Error())
			}
			sourcefiles = append(sourcefiles, sourcefile)
		}
	})

	// lexing
	timed("lexing phase", func() {
		for _, file := range sourcefiles {
			file.Tokens = lexer.Lex(file.Contents, file.Name)
		}
	})

	// parsing
	var parsedFiles []*parser.Module
	modules := make(map[string]*parser.Module)

	timed("parsing phase", func() {
		for _, file := range sourcefiles {
			parsedFiles = append(parsedFiles, parser.Parse(file, modules))
		}
	})

	// resolve
	timed("resolve phase", func() {
		// TODO: We're looping over a map, the order we get is thus random
		for _, module := range modules {
			res := &parser.Resolver{Module: module}
			res.Resolve(modules)
		}
	})

	// semantic analysis
	timed("semantic analysis phase", func() {
		// TODO: We're looping over a map, the order we get is thus random
		for _, module := range modules {
			sem := &parser.SemanticAnalyzer{Module: module}
			sem.Analyze(modules)
		}
	})

	// codegen
	if cg != "none" {
		var gen codegen.Codegen

		switch cg {
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName:   outputFile,
				CompilerArgs: ccArgs,
				OutputType:   outputType,
			}
		default:
			log.Error("main", util.Red("error: ")+"Invalid backend choice `"+cg+"`")
			os.Exit(1)
		}

		timed("codegen phase", func() {
			gen.Generate(parsedFiles, modules)
		})
	}

}

func timed(title string, fn func()) {
	log.Verboseln("main", util.TEXT_BOLD+util.TEXT_GREEN+"Started "+title+util.TEXT_RESET)
	start := time.Now()

	fn()

	duration := time.Since(start)
	log.Verboseln("main", util.TEXT_BOLD+util.TEXT_GREEN+"Ended "+title+util.TEXT_RESET+" (%.2fms)", float32(duration)/1000000)
}

func run(output string) {
	cmd := exec.Command("./" + output)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func docgen(input []string, dir string) {
	files, _ := parseFiles(input)

	gen := &doc.Docgen{
		Input: files,
		Dir:   dir,
	}

	gen.Generate()
}
