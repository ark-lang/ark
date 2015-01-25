#ifndef INK_LANG_H
#define INK_LANG_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "semantic.h"
#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "scanner.h"
#include "preprocessor.h"
#include "compiler.h"

/**
 * For handling command line
 * arguments
 */
typedef struct {
	char *argument;
	char *next_argument;
} argument;

/**
 * The core of inkc
 */
typedef struct {
	scanner *scanner;
	lexer *lexer;
	parser *parser;
	compiler *compiler;
	preprocessor *pproc;
    semantic *semantic;
	char *filename;
} inkc;

/**
 * Creates a new inkc instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of inkc
 */
inkc *create_inkc(int argc, char** argv);

/**
 * Start the inkc stuff
 * 
 * @param inkc instance to start
 */
void start_inkc(inkc *self);

/**
 * Runs the bytecode in the VM
 * @param self the inkc instance
 */
void run_vm_executable(inkc *self);

/**
 * Destroy the given inkc instance
 * 
 * @param inkc instance to destroy
 */
void destroy_inkc(inkc *self);

#endif // INK_LANG_H
