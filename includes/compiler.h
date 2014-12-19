#ifndef COMPILER_H
#define COMPILER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "j4vm.h"
#include "hashmap.h"

typedef struct {
	Vector *ast;
	JayforVM *vm;
	Hashmap *functions;
	int *bytecode;

	int initialBytecodeSize;
	int maxBytecodeSize;
	int currentNode;
	int currentInstruction;
	int globalCount;
} Compiler;

Compiler *createCompiler();

void appendInstruction(Compiler *self, int instr);

void generateFunctionCode(Compiler *self, FunctionNode *func);

int evaluateExpressionNode(Compiler *self, ExpressionNode *expr);

void consumeNode(Compiler *self);

void generateVariableDeclarationCode(Compiler *compiler, VariableDeclareNode *vdn);

void startCompiler(Compiler *compiler, Vector *ast);

void destroyCompiler(Compiler *compiler);

#endif // COMPILER_H