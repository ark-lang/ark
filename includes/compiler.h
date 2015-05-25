#ifndef __COMPILER_H
#define __COMPILER_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "semantic.h"
#include "LLVM/LLVMcodegen.h"
#include "hashmap.h"
#include "arguments.h"

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
	LLVMCodeGenerator *generatorLLVM;
	SemanticAnalyzer *semantic;
	Vector *sourceFiles;

    map_t arguments;
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
