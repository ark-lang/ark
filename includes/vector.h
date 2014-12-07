#ifndef VECTOR_H
#define VECTOR_H

#include <stdio.h>
#include <stdlib.h>

typedef void* VectorItem;

typedef struct {
	VectorItem *items;
	int maxSize;
	int size;
} Vector;

Vector *vectorCreate();

void vectorPushBack(Vector *vec, VectorItem item);

VectorItem vectorGetItem(Vector *vec, int index);

void vectorDestroy(Vector *vec);

#endif // VECTOR_H