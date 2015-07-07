package main

import "gopkg.in/alecthomas/kingpin.v2"

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

	buildCom     = app.Command("build", "Build an executable.")
	buildOutput  = buildCom.Flag("output", "Output binary name.").Short('o').Default("main").String()
	buildInputs  = newInputList(buildCom.Arg("input", "Ark source files."))
	buildCodegen = buildCom.Flag("codegen", "Codegen backend to use").Default("llvm").Enum("none", "llvm")
	buildStatic  = buildCom.Flag("static", "Pass the -static option to cc.").Bool()
	buildRun     = buildCom.Flag("run", "Run the executable.").Bool()
	buildAsm     = buildCom.Flag("output-assembly", "Stop compiling before assembling.").Short('S').Bool()

	docgenCom    = app.Command("docgen", "Generate documentation.")
	docgenDir    = docgenCom.Flag("dir", "Directory to place generated docs in.").Default("docgen").String()
	docgenInputs = newInputList(docgenCom.Arg("input", "Ark source files."))
)
