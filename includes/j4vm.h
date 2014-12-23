#ifndef jayfor_vm_H
#define jayfor_vm_H

#include <stdio.h>
#include <stdlib.h>
#include "stack.h"
#include "util.h"

typedef struct {
	int *bytecode;
	int *globals;
	int default_global_space;
	int instruction_pointer;
	int frame_pointer;
	bool running;
	Stack *stack;
} jayfor_vm;

typedef enum {
	ADD,
	SUB,
	MUL,
	DIV,
	MOD,
	POW,
	RET,
	PRINT,
	CALL,
	ICONST,
	LOAD,
	GLOAD,
	STORE,
	GSTORE,
	POP,
	HALT,
} instruction_set;

typedef struct {
    char name[6];
    int numOfArgs;
} instruction;

extern instruction debug_instructions[];

jayfor_vm *create_jayfor_vm();

void start_jayfor_vm(jayfor_vm *vm, int *bytecode);

void destroy_jayfor_vm(jayfor_vm *vm);

#endif // jayfor_vm_H