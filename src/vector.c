#include "vector.h"

Vector *vectorCreate() {
	Vector *vec = malloc(sizeof(*vec));
	if (!vec) {
		perror("malloc: failed to allocate memory for Vector");
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

void vectorPushBack(Vector *vec, VectorItem item) {
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

VectorItem vectorGetItem(Vector *vec, int index) {
	return vec->items[index];
}

void vectorDestroy(Vector *vec) {
	free(vec->items);
	free(vec);
}