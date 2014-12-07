#ifndef JAYFOR_H
#define JAYFOR_H

#include <stdio.h>
#include <stdlib.h>

#include "lexer.h"
#include "scanner.h"

typedef struct {
	Scanner *scanner;
	Lexer *lexer;
} Jayfor;

Jayfor *jayforCreate(int argc, char** argv);

void jayforStart(Jayfor *jayfor);

void jayforDestroy(Jayfor *jayfor);

#endif // JAYFOR_H