#ifndef ALLOY_LANG_H
#define ALLOY_LANG_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "compiler.h"

// flag for version of the alloy compiler
#define VERSION_ARG 		"-ver"

// flag for debug mode
#define DEBUG_MODE_ARG 		"-d"

// to define what compiler to compile the generated
// code with
#define COMPILER_ARG		"-compiler"

// flag that controls whether or not to keep
// the generated c files
#define OUTPUT_C_ARG		"-c"

// flag that shows the help dialog
#define HELP_ARG			"-h"

// flag for the executables name
#define OUTPUT_ARG			"-o"

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
	Compiler *compiler;
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

#endif // INK_LANG_H
