package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

		// build the files
		outputType := parseOutputType(*buildOutputType)
		build(*buildInputs, *buildOutput, *buildCodegen, ccArgs, outputType, *buildOptLevel)

		printFinishedMessage(startTime, buildCom.FullCommand(), len(*buildInputs))

		// attempt to run what we built
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
	log.Error("main", util.TEXT_RED+util.TEXT_BOLD+"error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_SETUP)
}

func parseFiles(inputs []string) ([]*parser.Module, *parser.ModuleLookup) {
	var modulesToRead []*parser.ModuleName
	for _, input := range inputs {
		if strings.ContainsAny(input, `\/. `) {
			setupErr("Invalid module name: %s", input)
		}

		modname := &parser.ModuleName{Parts: strings.Split(input, "::")}
		modulesToRead = append(modulesToRead, modname)
	}

	var parsedFiles []*parser.ParseTree
	moduleLookup := parser.NewModuleLookup("")
	depGraph := parser.NewDependencyGraph()

	log.Timed("read/lex/parse phase", "", func() {
		for i := 0; i < len(modulesToRead); i++ {
			modname := modulesToRead[i]
			actualFile := filepath.Join(*buildBasedir, modname.ToPath())

			fi, err := os.Stat(actualFile + ".ark")
			shouldBeDir := false
			if os.IsNotExist(err) {
				fi, err = os.Stat(actualFile)
				shouldBeDir = true
			}

			if err != nil {
				setupErr("%s", err.Error())
			}

			if !fi.IsDir() && shouldBeDir {
				setupErr("Expected file `%s` to be directory, was file.", actualFile)
			}

			// Check lookup
			ll := moduleLookup.Create(modname)
			if ll.Tree == nil {
				if shouldBeDir {
					// Setup dummy parse tree
					ll.Tree = &parser.ParseTree{}
					ll.Tree.Name = modname

					// Setup dummy module
					ll.Module = &parser.Module{}
					ll.Module.Path = actualFile
					ll.Module.Name = modname
					ll.Module.GlobalScope = parser.NewGlobalScope()

					// Check module children
					childFiles, err := ioutil.ReadDir(actualFile)
					if err != nil {
						setupErr("%s", err.Error())
					}

					for _, childFile := range childFiles {
						if strings.HasPrefix(childFile.Name(), ".") {
							continue
						}

						if childFile.IsDir() || strings.HasSuffix(childFile.Name(), ".ark") {
							childmodname := parser.JoinModuleName(modname, strings.TrimSuffix(childFile.Name(), ".ark"))
							modulesToRead = append(modulesToRead, childmodname)

							ll.Module.GlobalScope.UsedModules[childmodname.Last()] = moduleLookup.Create(childmodname)
						}
					}
				} else {
					// Read
					sourcefile, err := lexer.NewSourcefile(actualFile + ".ark")
					if err != nil {
						setupErr("%s", err.Error())
					}

					// Lex
					sourcefile.Tokens = lexer.Lex(sourcefile)

					// Parse
					parsedFile, deps := parser.Parse(sourcefile)
					parsedFile.Name = modname

					parsedFiles = append(parsedFiles, parsedFile)
					ll.Tree = parsedFile

					// Add dependencies to parse array
					for _, dep := range deps {
						depname := parser.NewModuleName(dep)
						modulesToRead = append(modulesToRead, depname)
						depGraph.AddDependency(modname, depname)
					}
				}
			}
		}
	})

	// Check for cyclic dependencies (in modules)
	log.Timed("cyclic dependency check", "", func() {
		errs := depGraph.DetectCycles()
		if len(errs) > 0 {
			log.Errorln("main", "error: Encountered cyclic dependecies:")
			for _, cycle := range errs {
				log.Errorln("main", "%s", cycle)
			}
			os.Exit(util.EXIT_FAILURE_SETUP)
		}
	})

	// construction
	var constructedModules []*parser.Module
	log.Timed("construction phase", "", func() {
		for _, file := range parsedFiles {
			constructedModules = append(constructedModules, parser.Construct(file, moduleLookup))
		}
	})

	return constructedModules, moduleLookup
}

func build(files []string, outputFile string, cg string, ccArgs []string, outputType LLVMCodegen.OutputType, optLevel int) {
	constructedModules, _ := parseFiles(files)

	// resolve
	log.Timed("resolve phase", "", func() {
		for _, module := range constructedModules {
			res := &parser.Resolver{Module: module}
			vis := parser.NewASTVisitor(res)
			vis.VisitModule(module)
		}
	})

	// type inference
	log.Timed("inference phase", "", func() {
		for _, module := range constructedModules {
			inf := &parser.TypeInferer{Module: module}
			inf.Infer()

			// Dump AST
			log.Debugln("main", "AST of module `%s`:", module.Name)
			for _, node := range module.Nodes {
				log.Debugln("main", "%s", node.String())
			}
			log.Debugln("main", "")
		}
	})

	// semantic analysis
	log.Timed("semantic analysis phase", "", func() {
		for _, module := range constructedModules {
			sem := semantic.NewSemanticAnalyzer(module, *buildOwnership, *ignoreUnused)
			vis := parser.NewASTVisitor(sem)
			vis.VisitModule(module)
			sem.Finalize()
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

		log.Timed("codegen phase", "", func() {
			gen.Generate(constructedModules)
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
