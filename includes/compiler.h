#ifndef __COMPILER_H
#define __COMPILER_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "semantic.h"
#include "C/codegen.h"
#ifdef ENABLE_LLVM
	#include "LLVM/LLVMcodegen.h"
#endif

#define DEBUG_MODE_ARG 		"-d"
#define HELP_ARG       		"-h"
#define VERBOSE_ARG    		"-v"
#define OUTPUT_ARG     		"-o"
#define OUTPUT_C_ARG   		"-c"
#define COMPILER_ARG   		"--compiler"
#define VERSION_ARG    		"--version"
#define IGNORE_MAIN_ARG		"--no-main"
#ifdef ENABLE_LLVM
	#define LLVM_ARG   		"--llvm"
#endif

/**
 * For handling command line
 * arguments
 */
typedef struct {
	char *argument;
	char *nextArgument;
} CommandLineArgument;

/**
 * The core of the compiler
 */
typedef struct {
	Lexer *lexer;
	Parser *parser;
	CCodeGenerator *generator;
#ifdef ENABLE_LLVM
	LLVMCodeGenerator *generatorLLVM;
#endif
	SemanticAnalyzer *semantic;
	Vector *sourceFiles;
} Compiler;

/**
 * Creates a new compiler instance
 * @argc number of arguments
 * @argv argument list
 * @return instance of the compiler
 */
Compiler *createCompiler(int argc, char** argv);

/**
 * Start the compiler
 * @param compiler instance to start
 */
void startCompiler(Compiler *self);

/**
 * Destroy the given compiler instance
 * @param compiler instance to destroy
 */
void destroyCompiler(Compiler *self);

#endif // __COMPILER_H
