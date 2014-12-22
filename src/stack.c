#include "stack.h"

Stack *create_stack() {
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

StackItem get_value_from_stack(Stack *stack, int index) {
	if (index > stack->stackPointer) {
		printf(KRED "error: could not retrieve value at index %d\n" KNRM, index);
		return NULL;
	} 
	return stack->items[index];
}

void push_to_stack_at_index(Stack *stack, StackItem item, int index) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (stack->stackPointer >= stack->defaultStackSize) {
		stack->defaultStackSize *= 2;
		if (DEBUG_MODE) printf(KYEL "stack size expanded to: %d\n" KNRM, stack->defaultStackSize);

		StackItem *tmp = realloc(stack->items, sizeof(*stack->items) * stack->defaultStackSize);
		if (!tmp) {
			perror("realloc: failed to allocate memory for stack items");
			exit(1);
		}
		else {
			stack->items = tmp;
		}
	}
	stack->items[index] = item;
}

void push_to_stack(Stack *stack, StackItem item) {
	push_to_stack_at_index(stack, item, ++stack->stackPointer);
}

StackItem pop_stack(Stack *stack) {
	if (stack->stackPointer < 0) {
		printf(KRED "error: cannot pop value from empty stack\n" KNRM);
		return NULL;
	}
	return stack->items[stack->stackPointer--];
}

void destroy_stack(Stack *stack) {
	free(stack);
}
