#ifndef jayfor_vm_H
#define jayfor_vm_H

#include <stdio.h>
#include <stdlib.h>
#include "stack.h"
#include "util.h"


/**
 * Virtual Machine
 * (needs a lot of work!)
 */
typedef struct {
	int *bytecode;
	int *globals;
	int *stack;

	int instruction_pointer;
	int frame_pointer;
	int stack_pointer;
	int default_global_space;
	int default_stack_size;

	bool running;
} jayfor_vm;

/* Instruction Set for the VM */
typedef enum {
	ADD, 
	SUB, 
	MUL, 
	DIV, 
	MOD,
	RET, 
	PRINT, 
	CALL, 
	ICONST, 
	LOAD, 
	GLOAD, 
	STORE, 
	GSTORE,
	ILT, 
	IEQ, 
	BRF, 
	BRT, 
	POP, 
	HALT,
} instruction_set;

/**
 * Instructions with names and
 * number of arguments -- for debugging
 */
typedef struct {
    char name[6];
    int number_of_args;
} instruction;

/**
 * Global instruction array
 */
extern instruction debug_instructions[];

/**
 * Create a new VM instance
 * @return the instance created
 */
jayfor_vm *create_jayfor_vm();

/**
 * Start the VM
 * @param vm the VM to start
 * @param bytecode the code to execute
 * @param bytecode_size how many instructions there are
 */
void start_jayfor_vm(jayfor_vm *vm, int *bytecode, int bytecode_size);

/**
 * Destroys the given VM
 * @param vm the vm to destroy
 */
void destroy_jayfor_vm(jayfor_vm *vm);

#endif // jayfor_vm_H