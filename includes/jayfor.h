#ifndef jayfor_H
#define jayfor_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "util.h"
#include "lexer.h"
#include "parser.h"
#include "scanner.h"
#include "compiler.h"
#include "j4vm.h"

/**
 * For handling command line
 * arguments
 */
typedef struct {
	char *argument;
	char *next_argument;
} argument;

/**
 * The core of jayfor
 */
typedef struct {
	scanner *scanner;
	Lexer *lexer;
	parser *parser;
	compiler *compiler;
	jayfor_vm *j4vm;

	char *filename;
} jayfor;

/**
 * Creates a new jayfor instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of jayfor
 */
jayfor *create_jayfor(int argc, char** argv);

/**
 * Start the jayfor interpreter
 * 
 * @param jayfor instance to start
 */
void start_jayfor(jayfor *jayfor);

/**
 * Runs the bytecode in the VM
 * @param self the jayfor instance
 */
void run_vm_executable(jayfor *self);

/**
 * Destroy the given jayfor instance
 * 
 * @param jayfor instance to destroy
 */
void destroy_jayfor(jayfor *jayfor);

#endif // jayfor_H