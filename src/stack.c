#include "stack.h"

Stack *create_stack() {
	Stack *stack = safe_malloc(sizeof(*stack));
	stack->default_stack_size = 32;
	stack->stack_pointer = -1;
	stack->items = safe_malloc(sizeof(*stack->items) * stack->default_stack_size);
	return stack;
}

StackItem get_value_from_stack(Stack *stack, int index) {
	if (index > stack->stack_pointer) {
		error_message("error: could not retrieve value at index %d\n", index);
		return NULL;
	} 
	return stack->items[index];
}

void push_to_stack_at_index(Stack *stack, StackItem item, int index) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (stack->stack_pointer >= stack->default_stack_size) {
		stack->default_stack_size *= 2;
		if (DEBUG_MODE) debug_message("stack size expanded to: %d\n", stack->default_stack_size);

		StackItem *tmp = realloc(stack->items, sizeof(*stack->items) * stack->default_stack_size);
		if (!tmp) {
			perror("realloc: failed to allocate memory for stack items");
			return;
		}
		else {
			stack->items = tmp;
		}
	}
	stack->items[index] = item;
}

void push_to_stack(Stack *stack, StackItem item) {
	push_to_stack_at_index(stack, item, ++stack->stack_pointer);
}

StackItem pop_stack(Stack *stack) {
	if (stack->stack_pointer < 0) {
		error_message("error: cannot pop value from empty stack\n");
		return NULL;
	}
	return stack->items[stack->stack_pointer--];
}

void destroy_stack(Stack *stack) {
	if (stack) {
		free(stack);
	}
}
