#include "stack.h"

Stack *createStack() {
	Stack *stack = malloc(sizeof(*stack));
	if (!stack) {
		perror("malloc: failed to allocate memory for stack");
		exit(1);
	}
	stack->defaultStackSize = 32;
	stack->stackPointer = -1;
	stack->items = malloc(sizeof(*stack->items) * stack->defaultStackSize);
	if (!stack->items) {
		perror("malloc: failed to allocate memory for stack items");
		exit(1);
	}
	return stack;
}

StackItem getValueFromStack(Stack *stack, int index) {
	if (index > stack->stackPointer) {
		printf("error: could not retrieve value at index %d\n", index);
		return NULL;
	} 
	return stack->items[index];
}

void pushToStackAtIndex(Stack *stack, StackItem item, int index) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (stack->stackPointer >= stack->defaultStackSize) {
		stack->defaultStackSize *= 2;
		stack->items = realloc(stack->items, sizeof(*stack->items) * stack->defaultStackSize);
		if (!stack->items) {
			perror("realloc: failed to allocate memory for stack items");
			exit(1);
		}
	}
	stack->items[index] = item;
}

void pushToStack(Stack *stack, StackItem item) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (stack->stackPointer >= stack->defaultStackSize) {
		stack->defaultStackSize *= 2;
		stack->items = realloc(stack->items, sizeof(*stack->items) * stack->defaultStackSize);
		if (!stack->items) {
			perror("realloc: failed to allocate memory for stack items");
			exit(1);
		}
	}
	stack->items[++stack->stackPointer] = item;
}

StackItem popStack(Stack *stack) {
	if (stack->stackPointer < 0) {
		printf("error: cannot pop value from empty stack\n");
		return NULL;
	}
	return stack->items[stack->stackPointer--];
}

void destroyStack(Stack *stack) {
	free(stack);
}
