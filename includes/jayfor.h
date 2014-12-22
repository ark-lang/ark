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
 * The core of jayfor
 */
typedef struct {
	Scanner *scanner;
	Lexer *lexer;
	parser *parser;
	compiler *compiler;
	jayfor_vm *j4vm;
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
 * Destroy the given jayfor instance
 * 
 * @param jayfor instance to destroy
 */
void destroy_jayfor(jayfor *jayfor);

#endif // jayfor_H