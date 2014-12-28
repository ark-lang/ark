#ifndef jayfor_vm_H
#define jayfor_vm_H

#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#include "util.h"

/** maximum number of locals TODO: make dynamic */
#define MAX_LOCAL_COUNT 10

/** maximum number of stack items TODO: make dynamic */
#define MAX_STACK_COUNT 10

/** OP CODES */
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
	HALT,
};

/** TYPES */
enum {
	type_object,
	type_number,
	type_string
};

/**
 * object properties
 */
typedef struct {
	char type;
	union {
		char* string;
		int number;
	} value;
	int ref_count;
} object;

/** default objects */
static object *true_object;
static object *false_object;
static object *null_object;

/**
 * The Jayfor VM (J4VM)
 */
typedef struct {
	int *instructions;
	object *stack[MAX_STACK_COUNT];
	object **stack_pointer;
	object *locals[MAX_LOCAL_COUNT];
	object *self;
} jayfor_vm;

/** garbage collection */

/**
 * Creates a new object
 * @return the object instance created
 */
object *create_object();

/**
 * Retains the given object
 * @param  obj the object to retain
 * @return the object retained
 */
object *retain_object(object *obj);

/**
 * Releases the given object
 * @param obj the object to release
 */
void release_object(object *obj);

/** virtual machine stuff */

object *get_current_stack_item(jayfor_vm *vm);

object *vm_pop(jayfor_vm *vm);

void vm_push(jayfor_vm *vm, object *obj);

object *get_local(jayfor_vm *vm);

object *get_local_at_index(jayfor_vm *vm, int index);

void set_local_at_index(jayfor_vm *vm, object *obj, int index);

void set_local(jayfor_vm *vm, object *obj);

/**
 * Creates a JVM installation
 * @return the JVM instance created
 */
jayfor_vm *create_jayfor_vm();

/**
 * Starts the given JVM
 * @param jvm          the JVM instance to start
 * @param instructions the list of instructions to execute
 */
void start_jayfor_vm(jayfor_vm *jvm, int *instructions);

/**
 * Destroys the given JVM
 * @param jvm the JVM to destroy
 */
void destroy_jayfor_vm(jayfor_vm *jvm);

/** helpers */

/**
 * Check if the given object is true
 * @param  obj the object to check
 * @return     if the object is true
 */
bool object_is_true(object *obj);

/**
 * Get the number value of an object
 * @param  obj the object to retrieve the value from
 * @return     the objects value as an integer
 */
int number_value(object *obj);

/**
 * Create a new number with the given value
 * @param  value the value of the number to create
 * @return       the new object
 */
object *create_number(int value);

/**
 * Creates a new string object
 * @param  value the value to give the string
 * @return       the string object created
 */
object *create_string(char *value);

#endif // jayfor_vm_H