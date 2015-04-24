#ifndef __STACK_H
#define __STACK_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

/**
 * typedef for void* so we can
 * change the type if need be
 */
typedef void* StackItem;

/**
 * Stack properties 
 */
typedef struct {
	StackItem *items;		// the items for the stack
	int stackPointer;		// points to the top of the stack
	int defaultStackSize;	// the initial size of the stack
} Stack;

/**
 * Create a new stack instance
 * @return the instance of the stack created
 */
Stack *createStack();

/**
 * Push a value to the stack
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void pushToStack(Stack *stack, StackItem item);

/**
 * Push a value to the stack at the given index
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void pushToStackAtIndex(Stack *stack, StackItem item, int index);

/**
 * Retrieve the value at the given index
 * 
 * @param stack the stack to retrieve from
 * @param index the index to check
 * @return the item at the given @{index}
 */
StackItem getStackItem(Stack *stack, int index);

/**
 * Pops value from top of stack
 * 
 * @param stack the stack to pop
 * @return the item we popped
 */
StackItem popStack(Stack *stack);

/**
 * Destroys the given stack
 * 
 * @param stack stack to destroy
 */
void destroyStack(Stack *stack);

#endif // __STACK_H
