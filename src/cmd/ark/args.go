package main

import (
	"github.com/ark-lang/ark/src/codegen/LLVMCodegen"
	"gopkg.in/alecthomas/kingpin.v2"
)

type inputList []string

func (i *inputList) String() string     { return "" }
func (i *inputList) IsCumulative() bool { return true }

func (i *inputList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func newInputList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*inputList)(target))
	return
}

var (
	app = kingpin.New("ark", "Compiler for the Ark programming language.").Version(VERSION).Author(AUTHOR)

	logLevel = app.Flag("loglevel", "Set the level of logging to show").Default("info").Enum("debug", "verbose", "info", "warning", "error")
	logTags  = app.Flag("logtags", "Which log tags to show").Default("all").String()

	buildCom        = app.Command("build", "Build an executable.")
	buildOutput     = buildCom.Flag("output", "Output binary name.").Short('o').Default("main").String()
	buildInputs     = newInputList(buildCom.Arg("input", "Ark source files."))
	buildCodegen    = buildCom.Flag("codegen", "Codegen backend to use").Default("llvm").Enum("none", "llvm")
	buildStatic     = buildCom.Flag("static", "Pass the -static option to cc.").Bool()
	buildRun        = buildCom.Flag("run", "Run the executable.").Bool()
	buildOutputType = buildCom.Flag("output-type", "Codegen backend to use").Default("executable").Enum("executable", "assembly", "object", "llvm-ir", "llvm-bc")
	buildOptLevel   = buildCom.Flag("opt-level", "LLVM optimization level").Short('O').Default("0").Int()

	docgenCom    = app.Command("docgen", "Generate documentation.")
	docgenDir    = docgenCom.Flag("dir", "Directory to place generated docs in.").Default("docgen").String()
	docgenInputs = newInputList(docgenCom.Arg("input", "Ark source files."))
)

func parseOutputType(name string) LLVMCodegen.OutputType {
	switch name {
	case "executable":
		return LLVMCodegen.OUTPUT_EXECUTABLE
	case "assembly":
		return LLVMCodegen.OUTPUT_ASSEMBLY
	case "object":
		return LLVMCodegen.OUTPUT_OBJECT
	case "llvm-ir":
		return LLVMCodegen.OUTPUT_LLVM_IR
	case "llvm-bc":
		return LLVMCodegen.OUTPUT_LLVM_BC
	default:
		panic("unimplemented output type")
	}
}
