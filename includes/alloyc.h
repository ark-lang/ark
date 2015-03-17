#ifndef ALLOY_LANG_H
#define ALLOY_LANG_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "compiler.h"

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
