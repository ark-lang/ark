#ifndef INTERPRETER_H
#define INTERPRETER_H

#include <stdio.h>
#include <stdlib.h>
#include "vector.h"

typedef struct {
	Vector *ast;
} Interpreter;

Interpreter *createInterpreter();

void startInterpreter(Vector *ast);

void destroyInterpreter(Interpreter *interpreter);

#endif // INTERPRETER_H