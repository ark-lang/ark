#ifndef ALLOY_LANG_H
#define ALLOY_LANG_H

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
 * The core of alloyc
 */
typedef struct {
	scanner *scanner;
	lexer *lexer;
	parser *parser;
	compiler *compiler;
	preprocessor *pproc;
    	semantic *semantic;
	char *filename;
} alloyc;

/**
 * Creates a new alloyc instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of alloyc
 */
alloyc *create_alloyc(int argc, char** argv);

/**
 * Start the alloyc stuff
 * 
 * @param alloyc instance to start
 */
void start_alloyc(alloyc *self);

/**
 * Runs the bytecode in the VM
 * @param self the alloyc instance
 */
void run_vm_executable(alloyc *self);

/**
 * Destroy the given alloyc instance
 * 
 * @param alloyc instance to destroy
 */
void destroy_alloyc(alloyc *self);

#endif // INK_LANG_H
