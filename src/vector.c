#include "vector.h"

Vector *createVector() {
	Vector *vec = safeMalloc(sizeof(*vec));
	vec->size = 0;
	vec->maxSize = 2;
	vec->items = safeMalloc(sizeof(*vec->items) * vec->maxSize);
	return vec;
}

void pushBackItem(Vector *vec, VectorItem item) {
	if (vec) {
		// much more efficient to reallocate exponentially,
		// instead of reallocating after adding an item
		if (vec->size >= vec->maxSize) {
			vec->maxSize *= 2;
			vec->items = realloc(vec->items, sizeof(*vec->items) * vec->maxSize);
			if (!vec->items) {
				perror("realloc: failed to allocate memory for vector contents");
				return;
			}
		}
		vec->items[vec->size++] = item;
	}
	else {
		errorMessage("Cannot push item to a null vector");
		return;
	}
}

VectorItem getVectorItem(Vector *vec, int index) {
	if (index > vec->size) {
		printf("index out of vector bounds, index: %d, size: %d\n", index, vec->size);
		return NULL;
	}
	return vec->items[index];
}

void destroyVector(Vector *vec) {
	if (vec) {
		free(vec->items);
		free(vec);
	}
	debugMessage("Destroyed Vector");
}
