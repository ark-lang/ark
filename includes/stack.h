#ifndef STACK_H
#define STACK_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

/**
 * typedef for void* so we can
 * change the type if need be
 */
typedef void* stack_item;

/**
 * Stack properties 
 */
typedef struct {
	stack_item *items;
	int stack_pointer;
	int default_stack_size;
} stack;

/**
 * Create a new stack instance
 * @return the instance of the stack created
 */
stack *create_stack();

/**
 * Push a value to the stack
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void push_to_stack(stack *stack, stack_item item);

/**
 * Push a value to the stack at the given index
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void push_to_stack_at_index(stack *stack, stack_item item, int index);

/**
 * Retrieve the value at the given index
 * 
 * @param stack the stack to retrieve from
 * @param index the index to check
 * @return the item at the given @{index}
 */
stack_item get_value_from_stack(stack *stack, int index);

/**
 * Pops value from top of stack
 * 
 * @param stack the stack to pop
 * @return the item we popped
 */
stack_item pop_stack(stack *stack);

/**
 * Destroys the given stack
 * 
 * @param stack stack to destroy
 */
void destroy_stack(stack *stack);

#endif // STACK_H