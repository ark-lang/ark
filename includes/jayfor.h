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

/**
 * The core of Jayfor
 */
typedef struct {
	Scanner *scanner;
	Lexer *lexer;
	Parser *parser;
	Compiler *compiler;
} Jayfor;

/**
 * Creates a new Jayfor instance
 * 
 * @argc number of arguments
 * @argv argument list
 * @return instance of Jayfor
 */
Jayfor *jayforCreate(int argc, char** argv);

/**
 * Start the Jayfor interpreter
 * 
 * @param jayfor instance to start
 */
void jayforStart(Jayfor *jayfor);

/**
 * Destroy the given Jayfor instance
 * 
 * @param jayfor instance to destroy
 */
void jayforDestroy(Jayfor *jayfor);

#endif // JAYFOR_H