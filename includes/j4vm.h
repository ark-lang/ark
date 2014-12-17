#ifndef JAYFORVM_H
#define JAYFORVM_H

#include <stdio.h>
#include <stdlib.h>
#include "stack.h"
#include "util.h"

typedef struct {
	int *bytecode;
	int *globals;
	int instructionPointer;
	int framePointer;
	bool running;
	Stack *stack;
} JayforVM;

typedef enum {
	ADD,
	SUB,
	MUL,
	RET,
	ICONST,
	LOAD,
	GLOAD,
	STORE,
	GSTORE,
	POP,
	HALT,
} InstructionSet;

JayforVM *createJayforVM();

void startJayforVM(JayforVM *vm, int *bytecode, int globalCount);

void destroyJayforVM(JayforVM *vm);

#endif // JAYFORVM_H