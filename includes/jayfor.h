#ifndef JAYFOR_H
#define JAYFOR_H

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
 * The core of Jayfor
 */
typedef struct {
	Scanner *scanner;
	Lexer *lexer;
	parser *parser;
	compiler *compiler;
	jayfor_vm *j4vm;
} Jayfor;

/**
 * Creates a new Jayfor instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of Jayfor
 */
Jayfor *create_jayfor(int argc, char** argv);

/**
 * Start the Jayfor interpreter
 * 
 * @param jayfor instance to start
 */
void start_jayfor(Jayfor *jayfor);

/**
 * Destroy the given Jayfor instance
 * 
 * @param jayfor instance to destroy
 */
void destroy_jayfor(Jayfor *jayfor);

#endif // JAYFOR_H