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
	return stack->items[index];
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
	return stack->items[stack->stackPointer--];
}

void destroyStack(Stack *stack) {
	free(stack);
}
