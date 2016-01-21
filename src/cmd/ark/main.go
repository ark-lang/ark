package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

		// build the files
		outputType := parseOutputType(*buildOutputType)
		build(*buildInputs, *buildOutput, *buildCodegen, outputType, *buildOptLevel)

		printFinishedMessage(startTime, buildCom.FullCommand(), len(*buildInputs))

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
	if len(inputs) != 1 {
		setupErr("Please specify only one file or module to build")
	}

	var modulesToRead []*parser.ModuleName
	var modules []*parser.Module
	moduleLookup := parser.NewModuleLookup("")
	depGraph := parser.NewDependencyGraph()

	input := inputs[0]
	if strings.HasSuffix(input, ".ark") {
		// Handle the special case of a single .ark file
		modname := &parser.ModuleName{Parts: []string{"__main"}}
		module := &parser.Module{
			Name:    modname,
			Dirpath: "",
		}
		moduleLookup.Create(modname).Module = module

		// Read
		sourcefile, err := lexer.NewSourcefile(input)
		if err != nil {
			setupErr("%s", err.Error())
		}

		// Lex
		sourcefile.Tokens = lexer.Lex(sourcefile)

		// Parse
		parsedFile, deps := parser.Parse(sourcefile)
		module.Trees = append(module.Trees, parsedFile)

		// Add dependencies to parse array
		for _, dep := range deps {
			depname := parser.NewModuleName(dep)
			modulesToRead = append(modulesToRead, depname)
			depGraph.AddDependency(modname, depname)
		}

		modules = append(modules, module)
	} else {
		if strings.ContainsAny(input, `\/. `) {
			setupErr("Invalid module name: %s", input)
		}

		modname := &parser.ModuleName{Parts: strings.Split(input, "::")}
		modulesToRead = append(modulesToRead, modname)
	}

	log.Timed("read/lex/parse phase", "", func() {
		for i := 0; i < len(modulesToRead); i++ {
			modname := modulesToRead[i]

			// Skip already loaded modules
			if _, err := moduleLookup.Get(modname); err == nil {
				continue
			}

			fi, dirpath := findModuleDir(*buildSearchpaths, modname.ToPath())
			if fi == nil {
				setupErr("Couldn't find module `%s`", modname)
			}

			if !fi.IsDir() {
				setupErr("Expected path `%s` to be directory, was file.", dirpath)
			}

			module := &parser.Module{
				Name:    modname,
				Dirpath: dirpath,
			}
			moduleLookup.Create(modname).Module = module

			// Check module children
			childFiles, err := ioutil.ReadDir(dirpath)
			if err != nil {
				setupErr("%s", err.Error())
			}

			for _, childFile := range childFiles {
				if strings.HasPrefix(childFile.Name(), ".") || !strings.HasSuffix(childFile.Name(), ".ark") {
					continue
				}

				actualFile := filepath.Join(dirpath, childFile.Name())

				// Read
				sourcefile, err := lexer.NewSourcefile(actualFile)
				if err != nil {
					setupErr("%s", err.Error())
				}

				// Lex
				sourcefile.Tokens = lexer.Lex(sourcefile)

				// Parse
				parsedFile, deps := parser.Parse(sourcefile)
				module.Trees = append(module.Trees, parsedFile)

				// Add dependencies to parse array
				for _, dep := range deps {
					depname := parser.NewModuleName(dep)
					modulesToRead = append(modulesToRead, depname)
					depGraph.AddDependency(modname, depname)
				}
			}

			modules = append(modules, module)
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
	log.Timed("construction phase", "", func() {
		for _, module := range modules {
			parser.Construct(module, moduleLookup)

		}
	})

	return modules, moduleLookup
}

func findModuleDir(searchPaths []string, modulePath string) (fi os.FileInfo, path string) {
	for _, searchPath := range searchPaths {
		path := filepath.Join(searchPath, modulePath)
		fi, err := os.Stat(path)
		if err == nil {
			return fi, path
		}

		if !os.IsNotExist(err) {
			setupErr("%s\n", err)
		}
	}

	return nil, ""
}

func build(files []string, outputFile string, cg string, outputType LLVMCodegen.OutputType, optLevel int) {
	constructedModules, moduleLookup := parseFiles(files)

	// resolve

	hasMainFunc := false
	log.Timed("resolve phase", "", func() {
		for _, module := range constructedModules {
			parser.Resolve(module, moduleLookup)

			// Use module scope to check for main function
			mainIdent := module.ModScope.GetIdent(parser.UnresolvedName{Name: "main"})
			if mainIdent != nil && mainIdent.Type == parser.IDENT_FUNCTION && mainIdent.Public {
				hasMainFunc = true
			}
		}
	})

	// and here we check if we should
	// bother continuing any further...
	if !hasMainFunc {
		log.Error("main", util.Red("error: ")+"main function not found\n")
		os.Exit(1)
	}

	// type inference
	log.Timed("inference phase", "", func() {
		for _, module := range constructedModules {
			for _, submod := range module.Parts {
				inf := &parser.TypeInferer{Submodule: submod}
				inf.Infer()

				// Dump AST
				log.Debugln("main", "AST of submodule `%s/%s`:", module.Name, submod.File.Name)
				for _, node := range submod.Nodes {
					log.Debugln("main", "%s", node.String())
				}
				log.Debugln("main", "")
			}
		}
	})

	// semantic analysis
	log.Timed("semantic analysis phase", "", func() {
		for _, module := range constructedModules {
			for _, submod := range module.Parts {
				sem := semantic.NewSemanticAnalyzer(submod, *buildOwnership, *ignoreUnused)
				vis := parser.NewASTVisitor(sem)
				vis.VisitSubmodule(submod)
				sem.Finalize()
			}
		}
	})

	// codegen
	if cg != "none" {
		var gen codegen.Codegen

		switch cg {
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName: outputFile,
				OutputType: outputType,
				OptLevel:   optLevel,
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

func docgen(input []string, dir string) {
	constructedModules, _ := parseFiles(input)

	gen := &doc.Docgen{
		Input: constructedModules,
		Dir:   dir,
	}

	gen.Generate()
}
