#ifndef __ALLOY_LANG_H
#define __ALLOY_LANG_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "semantic.h"
#include "C/codegen.h"

#define DEBUG_MODE_ARG 		"-d"
#define HELP_ARG			"-h"
#define VERBOSE_ARG			"-v"
#define OUTPUT_ARG			"-o"
#define OUTPUT_C_ARG		"-c"
#define COMPILER_ARG		"--compiler"
#define VERSION_ARG 		"--version"

/**
 * For handling command line
 * arguments
 */
typedef struct {
	char *argument;
	char *nextArgument;
} CommandLineArgument;

/**
 * The core of alloyc
 */
typedef struct {
	Lexer *lexer;
	Parser *parser;
	CCodeGenerator *generator;
	SemanticAnalyzer *semantic;
	Vector *sourceFiles;
} AlloyCompiler;

/**
 * Creates a new alloyc instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of alloyc
 */
AlloyCompiler *createAlloyCompiler(int argc, char** argv);

/**
 * Start the alloyc stuff
 * 
 * @param alloyc instance to start
 */
void startAlloyCompiler(AlloyCompiler *self);

/**
 * Destroy the given alloyc instance
 * 
 * @param alloyc instance to destroy
 */
void destroyAlloyCompiler(AlloyCompiler *self);

#endif // __ALLOY_LANG_H
