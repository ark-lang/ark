package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ark-lang/ark/src/codegen"
	"github.com/ark-lang/ark/src/codegen/LLVMCodegen"
	"github.com/ark-lang/ark/src/doc"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/semantic"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
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
		build(*buildInputs, *buildOutput, *buildCodegen, ccArgs, outputType, *buildOptLevel)
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
	// read source files
	var sourcefiles []*lexer.Sourcefile
	log.Timed("reading sourcefiles", func() {
		for _, file := range files {
			sourcefile, err := lexer.NewSourcefile(file)
			if err != nil {
				setupErr("%s", err.Error())
			}
			sourcefiles = append(sourcefiles, sourcefile)
		}
	})

	// lexing
	log.Timed("lexing phase", func() {
		for _, file := range sourcefiles {
			file.Tokens = lexer.Lex(file)
		}
	})

	// parsing
	var parsedFiles []*parser.ParseTree
	parsedFileMap := make(map[string]*parser.ParseTree)
	log.Timed("parsing phase", func() {
		for _, file := range sourcefiles {
			parsedFile := parser.Parse(file)
			parsedFiles = append(parsedFiles, parsedFile)
			parsedFileMap[parsedFile.Source.Name] = parsedFile
		}
	})

	// construction
	var constructedModules []*parser.Module
	modules := make(map[string]*parser.Module)
	log.Timed("construction phase", func() {
		for _, file := range parsedFiles {
			constructedModules = append(constructedModules, parser.Construct(file, parsedFileMap, modules))
		}
	})

	return constructedModules, modules
}

func build(files []string, outputFile string, cg string, ccArgs []string, outputType LLVMCodegen.OutputType, optLevel int) {
	constructedModules, modules := parseFiles(files)

	// resolve
	log.Timed("resolve phase", func() {
		// TODO: We're looping over a map, the order we get is thus random
		for _, module := range modules {
			res := &parser.Resolver{Module: module}
			vis := parser.NewASTVisitor(res)
			vis.VisitModule(module)
		}
	})

	// type inference
	log.Timed("inference phase", func() {
		// TODO: We're looping over a map, the order we get is thus random
		for _, module := range modules {
			inf := &parser.TypeInferer{Module: module}
			inf.Infer(modules)

			// Dump AST
			log.Debugln("main", "AST of module `%s`:", module.Name)
			for _, node := range module.Nodes {
				log.Debugln("main", "%s", node.String())
			}
		}
	})

	// semantic analysis
	log.Timed("semantic analysis phase", func() {
		// TODO: We're looping over a map, the order we get is thus random
		for _, module := range modules {
			sem := semantic.NewSemanticAnalyzer(module)
			vis := parser.NewASTVisitor(sem)
			vis.VisitModule(module)
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
				OptLevel:     optLevel,
			}
		default:
			log.Error("main", util.Red("error: ")+"Invalid backend choice `"+cg+"`")
			os.Exit(1)
		}

		log.Timed("codegen phase", func() {
			gen.Generate(constructedModules, modules)
		})
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
	constructedModules, _ := parseFiles(input)

	gen := &doc.Docgen{
		Input: constructedModules,
		Dir:   dir,
	}

	gen.Generate()
}
