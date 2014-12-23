#ifndef STACK_H
#define STACK_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

/**
 * Quick Stack implementation
 */

/**
 * typedef for void* so we can
 * change the type if need be
 */
typedef void* StackItem;

/**
 * Stack properties 
 */
typedef struct {
	StackItem *items;
	int stack_pointer;
	int default_stack_size;
} Stack;

/**
 * Create a new stack instance
 * @return the instance of the stack created
 */
Stack *create_stack();

/**
 * Push a value to the stack
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void push_to_stack(Stack *stack, StackItem item);

/**
 * Push a value to the stack at the given index
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void push_to_stack_at_index(Stack *stack, StackItem item, int index);

/**
 * Retrieve the value at the given index
 * 
 * @param stack the stack to retrieve from
 * @param index the index to check
 * @return the item at the given @{index}
 */
StackItem get_value_from_stack(Stack *stack, int index);

/**
 * Pops value from top of stack
 * 
 * @param stack the stack to pop
 * @return the item we popped
 */
StackItem pop_stack(Stack *stack);

/**
 * Destroys the given stack
 * 
 * @param stack stack to destroy
 */
void destroy_stack(Stack *stack);

#endif // STACK_H