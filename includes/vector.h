#ifndef vector_H
#define vector_H

#include <stdio.h>
#include <stdlib.h>
#include "util.h"

/**
 * vector_item typedef so we can change
 * the data type if need be
 */
typedef void* vector_item;

/**
 * Properties of a vector
 */
typedef struct {
	vector_item *items;
	int max_size;
	int size;
} vector;

/**
 * Creates a new instance of vector
 * 
 * @return the vector instance
 */
vector* create_vector();

/**
 * Push a vector_item into the given vector
 * 
 * @param vec the vector to push to
 * @param vector_item the item to push
 */
void push_back_item(vector *vec, vector_item item);

/**
 * Retrieves the vector_item at the given index in
 * the given vector
 * 
 * @param vec vector the retrieve item from
 * @param index the place to check
 * @return the vector Item
 */
vector_item get_vector_item(vector *vec, int index);

/**
 * Destroys the given vector
 * 
 * @param vec the vector to destroy
 */
void destroy_vector(vector *vec);

#endif // vector_H
