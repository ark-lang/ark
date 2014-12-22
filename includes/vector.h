#ifndef vector_H
#define vector_H

#include <stdio.h>
#include <stdlib.h>

/**
 * vectorItem typedef so we can change
 * the data type if need be
 */
typedef void* vectorItem;

/**
 * Properties of a vector
 */
typedef struct {
	vectorItem *items;
	int maxSize;
	int size;
} vector;

/**
 * Creates a new instance of vector
 * 
 * @return the vector instance
 */
vector *create_vector();

/**
 * Push a vectorItem into the given vector
 * 
 * @param vec the vector to push to
 * @param vectorItem the item to push
 */
void push_back_item(vector *vec, vectorItem item);

/**
 * Retrieves the vectorItem at the given index in
 * the given vector
 * 
 * @param vec vector the retrieve item from
 * @param index the place to check
 * @return the vector Item
 */
vectorItem get_vector_item(vector *vec, int index);

/**
 * Destroys the given vector
 * 
 * @param vec the vector to destroy
 */
void destroy_vector(vector *vec);

#endif // vector_H