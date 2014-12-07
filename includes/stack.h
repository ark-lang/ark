#ifndef STACK_H
#define STACK_H

#include <stdio.h>
#include <stdlib.h>

typedef void* StackItem;

typedef struct {
	StackItem *items;
	int stackPointer;
	int defaultStackSize;
} Stack;

Stack *stackCreate();

void stackPush(Stack *stack, StackItem item);

StackItem stackPop(Stack *stack);

void stackDestroy(Stack *stack);

#endif // STACK_H