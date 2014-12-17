#ifndef VECTOR_H
#define VECTOR_H

#include <stdio.h>
#include <stdlib.h>

/**
 * VectorItem typedef so we can change
 * the data type if need be
 */
typedef void* VectorItem;

/**
 * Properties of a Vector
 */
typedef struct {
	VectorItem *items;
	int maxSize;
	int size;
} Vector;

/**
 * Creates a new instance of Vector
 * 
 * @return the Vector instance
 */
Vector *createVector();

/**
 * Push a VectorItem into the given Vector
 * 
 * @param vec the vector to push to
 * @param VectorItem the item to push
 */
void pushBackVectorItem(Vector *vec, VectorItem item);

/**
 * Retrieves the VectorItem at the given index in
 * the given Vector
 * 
 * @param vec vector the retrieve item from
 * @param index the place to check
 * @return the Vector Item
 */
VectorItem getItemFromVector(Vector *vec, int index);

/**
 * Destroys the given Vector
 * 
 * @param vec the vector to destroy
 */
void destroyVector(Vector *vec);

#endif // VECTOR_H