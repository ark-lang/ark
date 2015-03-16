#ifndef hashmap_H
#define hashmap_H

#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#include "util.h"

typedef struct {
	char *key;
	void *val;
	size_t len;
} HashmapEntry;

typedef struct {
	int size;
	HashmapEntry *entries;
} HashmapField;

typedef struct {
	int size;
	HashmapField *fields;
} Hashmap;

/**
 * Creates a new Hashmap
 * @param int the size of the hashmap
 * @return the hashmap instance
 */
Hashmap *createHashmap(int);

/**
 * Destroys the given hashmap
 * @param map the hashmap to destroy
 */
void destroyHashmap(Hashmap *map);

/**
 * Get the value at the key specified
 * @param map the map to search
 * @param str the key to search for
 * @return the value in the map at the key
 */
void *getValueAtKey(Hashmap *map, char* str);

/**
 * Sets the value at the given key
 * @param map the map to set the value in
 * @param str the key to place the value
 * @param data the data to store
 */
void setValueAtKey(Hashmap *map, char* str, void* data);

#endif // hashmap_H
