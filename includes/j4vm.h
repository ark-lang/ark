#ifndef J4VM_H
#define J4VM_H

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

JayforVM *createJ4VM();

void startJ4VM(JayforVM *vm, int *bytecode);

void destroyJ4VM(JayforVM *vm);

#endif // J4VM_H