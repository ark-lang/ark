#ifndef COMPILER_H
#define COMPILER_H

#include <stdio.h>
#include <stdlib.h>

#include "vector.h"
#include "j4vm.h"

typedef struct {
	Vector *ast;
	JayforVM *vm;
} Compiler;

Compiler *createCompiler();

void startCompiler(Compiler *compiler, Vector *ast);

void destroyCompiler(Compiler *compiler);

#endif // COMPILER_H