#ifndef jayfor_vm_H
#define jayfor_vm_H

#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#include "util.h"

#define MAX_LOCAL_COUNT 10
#define MAX_STACK_COUNT 10
#define STACK_PUSH(JVM, VALUE) do {															\
							assert(JVM->stack_pointer - JVM->stack < MAX_STACK_COUNT);	\
							*(++JVM->stack_pointer) = (VALUE);												\
							retain_object(*JVM->stack_pointer);													\
						  } while (0);

#define STACK_POP(JVM)	(*JVM->stack_pointer--)

typedef unsigned char byte;

enum {
	CALL,
	PUSH_NUMBER,
	PUSH_STRING,
	PUSH_SELF,
	PUSH_NULL,
	PUSH_BOOL,
	GET_LOCAL,
	SET_LOCAL,
	JUMP_UNLESS,
	JUMP,
	ADD,
	RETURN,
};

enum {
	type_object,
	type_number,
	type_string
};

typedef struct {
	char type;
	union {
		char* string;
		int number;
	} value;
	int ref_count;
} object;

static object *true_object;
static object *false_object;
static object *null_object;

typedef struct {
	int *instructions;
	object *stack[MAX_STACK_COUNT];
	object **stack_pointer;
	object *locals[MAX_LOCAL_COUNT];
	object *self;
} jayfor_vm;

/** garbage collection */

object *create_object();

object *retain_object(object *obj);

void release_object(object *obj);

/** virtual machine stuff */

jayfor_vm *create_jayfor_vm();

void start_jayfor_vm(jayfor_vm *jvm, int *instructions);

void destroy_jayfor_vm(jayfor_vm *jvm);

/** helpers */

bool object_is_true(object *obj);

int number_value(object *obj);

object *create_number(int value);

object *create_string(char *value);

object *call(object *obj, char *message, object *argv[], int argc);

#endif // jayfor_vm_H