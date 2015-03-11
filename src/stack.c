#include "stack.h"

stack *create_stack() {
	stack *stack = safe_malloc(sizeof(*stack));
	stack->default_stack_size = 32;
	stack->stack_pointer = -1;
	stack->items = safe_malloc(sizeof(*stack->items) * stack->default_stack_size);
	return stack;
}

stack_item get_value_from_stack(stack *stack, int index) {
	if (index > stack->stack_pointer) {
		error_message("error: could not retrieve value at index %d\n", index);
		return NULL;
	} 
	return stack->items[index];
}

void push_to_stack_at_index(stack *stack, stack_item item, int index) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (stack->stack_pointer >= stack->default_stack_size) {
		stack->default_stack_size *= 2;
		if (DEBUG_MODE) debug_message("stack size expanded to: %d\n", stack->default_stack_size);

		stack_item *tmp = realloc(stack->items, sizeof(*stack->items) * stack->default_stack_size);
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

void push_to_stack(stack *stack, stack_item item) {
	push_to_stack_at_index(stack, item, ++stack->stack_pointer);
}

stack_item pop_stack(stack *stack) {
	if (stack->stack_pointer < 0) {
		error_message("error: cannot pop value from empty stack\n");
		return NULL;
	}
	return stack->items[stack->stack_pointer--];
}

void destroy_stack(stack *stack) {
	if (stack) {
		free(stack);
	}
}
