#include "vector.h"

vector *create_vector() {
	vector *vec = safe_malloc(sizeof(*vec));
	vec->size = 0;
	vec->max_size = 2;
	vec->items = safe_malloc(sizeof(*vec->items) * vec->max_size);
	return vec;
}

void push_back_item(vector *vec, vector_item item) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (vec->size >= vec->max_size) {
		vec->max_size *= 2;
		vec->items = realloc(vec->items, sizeof(*vec->items) * vec->max_size);
		if (!vec->items) {
			perror("realloc: failed to allocate memory for vector contents");
			exit(1);
		}
	}
	vec->items[vec->size++] = item;
}

vector_item get_vector_item(vector *vec, int index) {
	if (index > vec->size) {
		printf("index out of vector bounds, index: %d, size: %d\n", index, vec->size);
		exit(1);
	}
	return vec->items[index];
}

void destroy_vector(vector *vec) {
	if (vec) {
		if (vec->items) {
			free(vec->items);
		}
		free(vec);
	}
}
