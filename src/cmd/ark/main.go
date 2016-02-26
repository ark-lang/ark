package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ark-lang/ark/src/ast"
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

	context := NewContext()

	switch command {
	case buildCom.FullCommand():
		if *buildInput == "" {
			setupErr("No input files passed.")
		}

		context.Searchpaths = *buildSearchpaths
		context.Input = *buildInput

		outputType, err := codegen.ParseOutputType(*buildOutputType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// build the files
		context.Build(*buildOutput, outputType, *buildCodegen, *buildOptLevel)

		printFinishedMessage(startTime, buildCom.FullCommand(), 1)

	case docgenCom.FullCommand():
		context.Searchpaths = *docgenSearchpaths
		context.Input = *docgenInput
		context.Docgen(*docgenDir)

		printFinishedMessage(startTime, docgenCom.FullCommand(), 1)
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

type Context struct {
	Searchpaths []string

	Input string

	moduleLookup *ast.ModuleLookup
	depGraph     *ast.DependencyGraph
	modules      []*ast.Module

	modulesToRead []*ast.ModuleName
}

func NewContext() *Context {
	res := &Context{
		moduleLookup: ast.NewModuleLookup(""),
		depGraph:     ast.NewDependencyGraph(),
	}
	return res
}

func (v *Context) Build(output string, outputType codegen.OutputType, usedCodegen string, optLevel int) {
	// Start by loading the runtime
	runtimeModule := LoadRuntime()

	// Parse the passed files
	v.parseFiles()

	// resolve
	hasMainFunc := false
	log.Timed("resolve phase", "", func() {
		for _, module := range v.modules {
			ast.Resolve(module, v.moduleLookup)

			// Use module scope to check for main function
			mainIdent := module.ModScope.GetIdent(ast.UnresolvedName{Name: "main"})
			if mainIdent != nil && mainIdent.Type == ast.IDENT_FUNCTION && mainIdent.Public {
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
		for _, module := range v.modules {
			for _, submod := range module.Parts {
				ast.Infer(submod)

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
		for _, module := range v.modules {
			for _, submod := range module.Parts {
				sem := semantic.NewSemanticAnalyzer(submod, *buildOwnership, *ignoreUnused)
				vis := ast.NewASTVisitor(sem)
				vis.VisitSubmodule(submod)
				sem.Finalize()
			}
		}
	})

	// codegen
	if usedCodegen != "none" {
		var gen codegen.Codegen

		switch usedCodegen {
		case "llvm":
			gen = &LLVMCodegen.Codegen{
				OutputName: output,
				OutputType: outputType,
				OptLevel:   optLevel,
			}
		default:
			log.Error("main", util.Red("error: ")+"Invalid backend choice `"+usedCodegen+"`")
			os.Exit(1)
		}

		log.Timed("codegen phase", "", func() {
			mods := v.modules
			if runtimeModule != nil {
				mods = append(mods, runtimeModule)
			}
			gen.Generate(mods)
		})
	}
}

func (v *Context) Docgen(dir string) {
	v.parseFiles()

	gen := &doc.Docgen{
		Input: v.modules,
		Dir:   dir,
	}

	gen.Generate()
}

func (v *Context) parseFiles() {
	if strings.HasSuffix(v.Input, ".ark") {
		// Handle the special case of a single .ark file
		modname := &ast.ModuleName{Parts: []string{"__main"}}
		module := &ast.Module{
			Name:    modname,
			Dirpath: "",
		}
		v.moduleLookup.Create(modname).Module = module

		v.parseFile(v.Input, module)

		v.modules = append(v.modules, module)
	} else {
		if strings.ContainsAny(v.Input, `\/. `) {
			setupErr("Invalid module name: %s", v.Input)
		}

		modname := &ast.ModuleName{Parts: strings.Split(v.Input, "::")}
		v.modulesToRead = append(v.modulesToRead, modname)
	}

	log.Timed("read/lex/parse phase", "", func() {
		for i := 0; i < len(v.modulesToRead); i++ {
			modname := v.
				modulesToRead[i]

			// Skip already loaded modules
			if _, err := v.moduleLookup.Get(modname); err == nil {
				continue
			}

			fi, dirpath, err := v.findModuleDir(modname.ToPath())
			if err != nil {
				setupErr("Couldn't find module `%s`: %s", modname, err)
			}

			if !fi.IsDir() {
				setupErr("Expected path `%s` to be directory, was file.", dirpath)
			}

			module := &ast.Module{
				Name:    modname,
				Dirpath: dirpath,
			}
			v.moduleLookup.Create(modname).Module = module

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
				v.parseFile(actualFile, module)
			}

			v.modules = append(v.modules, module)
		}
	})

	// Check for cyclic dependencies (in modules)
	log.Timed("cyclic dependency check", "", func() {
		errs := v.depGraph.DetectCycles()
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
		for _, module := range v.modules {
			ast.Construct(module, v.moduleLookup)
		}
	})
}

func (v *Context) parseFile(path string, module *ast.Module) {
	// Read
	sourcefile, err := lexer.NewSourcefile(path)
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
		depname := ast.NewModuleName(dep)
		v.modulesToRead = append(v.modulesToRead, depname)
		v.depGraph.AddDependency(module.Name, depname)

		if _, _, err := v.findModuleDir(depname.ToPath()); err != nil {
			log.Errorln("main", "%s [%s:%d:%d] Couldn't find module `%s`", util.Red("error:"),
				dep.Where().Filename, dep.Where().StartLine, dep.Where().EndLine,
				depname.String())
			log.Errorln("main", "%s", sourcefile.MarkSpan(dep.Where()))
			os.Exit(1)
		}
	}
}

func (v *Context) findModuleDir(modulePath string) (fi os.FileInfo, path string, err error) {
	for _, searchPath := range v.Searchpaths {
		path := filepath.Join(searchPath, modulePath)
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		return fi, path, nil

	}

	return nil, "", fmt.Errorf("ark: Unable to find module `%s`", path)
}
