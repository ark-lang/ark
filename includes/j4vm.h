#ifndef jayfor_vm_H
#define jayfor_vm_H

#include <stdio.h>
#include <stdlib.h>
#include <math.h>

#include "util.h"

typedef enum {
	TYPE_FLOAT, TYPE_INT
} object_data_type;

typedef struct {
	int data;
	object_data_type type;
} object;

/**
 * Virtual Machine
 * (needs a lot of work!)
 */
typedef struct {
	int *bytecode;
	object *globals;
	object *stack;

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
	FCONST,
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
 * Convert a float to IEEE 754 int bits
 * @param x the float to convert
 */
int float_to_int_bits(float x);

/**
 * Convert integer bits to a float IEEE 754
 * @param x the int bits to convert
 */
float int_bits_to_float(int x);

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