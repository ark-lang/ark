#include "vector.h"

vector *create_vector() {
	vector *vec = malloc(sizeof(*vec));
	if (!vec) {
		perror("malloc: failed to allocate memory for vector");
		exit(1);
	}
	vec->size = 0;
	vec->maxSize = 2;
	vec->items = malloc(sizeof(*vec->items) * vec->maxSize);
	if (!vec->items) {
		perror("malloc: failed to allocate memory for vector contents");
		exit(1);
	}
	return vec;
}

void push_back_item(vector *vec, vectorItem item) {
	// much more efficient to reallocate exponentially,
	// instead of reallocating after adding an item
	if (vec->size >= vec->maxSize) {
		vec->maxSize *= 2;
		vec->items = realloc(vec->items, sizeof(*vec->items) * vec->maxSize);
		if (!vec->items) {
			perror("realloc: failed to allocate memory for vector contents");
			exit(1);
		}
	}
	vec->items[vec->size++] = item;
}

vectorItem get_vector_item(vector *vec, int index) {
	if (index > vec->size) {
		printf("index out of vector bounds, index: %d, size: %d\n", index, vec->size);
		exit(1);
	}
	return vec->items[index];
}

void destroy_vector(vector *vec) {
	free(vec->items);
	free(vec);
}