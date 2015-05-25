#include "vector.h"

Vector *createVector(int type) {
	Vector *vec = safeMalloc(sizeof(*vec));
	vec->size = 0;
	vec->maxSize = 2;
	vec->items = safeMalloc(sizeof(*vec->items) * vec->maxSize);
	vec->type = type;
	return vec;
}

void pushBackItem(Vector *vec, VectorItem item) {
	if (vec) {
		// much more efficient to reallocate exponentially,
		// instead of reallocating after adding an item
		if (vec->size >= vec->maxSize) {
			vec->maxSize = vec->type == VECTOR_LINEAR ? vec->maxSize + 1 : vec->maxSize * 2;
			vec->items = realloc(vec->items, sizeof(*vec->items) * vec->maxSize);
			if (!vec->items) {
				verboseModeMessage("Failed to allocate memory for vector contents");
				return;
			}
		}
		vec->items[vec->size++] = item;
	}
	else {
		verboseModeMessage("Cannot push item to a null vector");
		return;
	}
}

VectorItem getVectorItem(Vector *vec, int index) {
	if (index > vec->size) {
		verboseModeMessage("Index out of vector bounds, index: %d, size: %d", index, vec->size);
		return false;
	}
	return vec->items[index];
}

VectorItem getVectorTop(Vector *vec) {
	return getVectorItem(vec, vec->size);
}

void destroyVector(Vector *vec) {
	free(vec->items);
	free(vec);
	verboseModeMessage("Destroyed vector");
}
