#ifndef STACK_H
#define STACK_H

#include <stdio.h>
#include <stdlib.h>

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
	int stackPointer;
	int defaultStackSize;
} Stack;

/**
 * Create a new stack instance
 * @return the instance of the stack created
 */
Stack *stackCreate();

/**
 * Push a value to the stack
 * 
 * @param stack stack to push to
 * @param item item to push to
 */
void stackPush(Stack *stack, StackItem item);

/**
 * Pops value from top of stack
 * 
 * @param stack the stack to pop
 * @return the item we popped
 */
StackItem stackPop(Stack *stack);

/**
 * Destroys the given stack
 * 
 * @param stack stack to destroy
 */
void stackDestroy(Stack *stack);

#endif // STACK_H