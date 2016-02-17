package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("ark", "Compiler for the Ark programming language.").Version(VERSION).Author(AUTHOR)

	logLevel = app.Flag("loglevel", "Set the level of logging to show").Default("info").Enum("debug", "verbose", "info", "warning", "error")
	logTags  = app.Flag("logtags", "Which log tags to show").Default("all").String()

	buildCom         = app.Command("build", "Build an executable.")
	buildOutput      = buildCom.Flag("output", "Output binary name.").Short('o').Default("main").String()
	buildSearchpaths = buildCom.Flag("searchpaths", "Paths to search for used modules if not found in base directory").Short('I').Strings()
	buildInput       = buildCom.Arg("input", "Ark source file or package").String()
	buildCodegen     = buildCom.Flag("codegen", "Codegen backend to use").Default("llvm").Enum("none", "llvm")
	buildOutputType  = buildCom.Flag("output-type", "The format to produce after code generation").Default("executable").Enum("executable", "assembly", "object", "llvm-ir")
	buildOptLevel    = buildCom.Flag("opt-level", "LLVM optimization level").Short('O').Default("0").Int()
	buildOwnership   = buildCom.Flag("ownership", "Do ownership checks").Bool()
	ignoreUnused     = buildCom.Flag("unused", "Do not error on unused declarations").Bool()

	docgenCom         = app.Command("docgen", "Generate documentation.")
	docgenDir         = docgenCom.Flag("dir", "Directory to place generated docs in.").Default("docgen").String()
	docgenInput       = docgenCom.Arg("input", "Ark source file or package").String()
	docgenSearchpaths = docgenCom.Flag("searchpaths", "Paths to search for used modules if not found in base directory").Short('I').Strings()
)
