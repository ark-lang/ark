#ifndef hashmap_H
#define hashmap_H

#include <stdlib.h>
#include <string.h>

typedef struct {
	char *key;
	void *val;
	size_t len;
} hashmap_entry;

typedef struct {
	int size;
	hashmap_entry *entries;
} hashmap_field;

typedef struct {
	int size;
	hashmap_field *fields;
} hashmap;

/**
 * Creates a new Hashmap
 * @param int the size of the hashmap
 * @return the hashmap instance
 */
hashmap *create_hashmap(int);

/**
 * Destroys the given hashmap
 * @param map the hashmap to destroy
 */
void destroy_hashmap(hashmap *map);

/**
 * Get the value at the key specified
 * @param map the map to search
 * @param str the key to search for
 * @return the value in the map at the key
 */
void *get_value_at_key(hashmap *map, char* str);

/**
 * Sets the value at the given key
 * @param map the map to set the value in
 * @param str the key to place the value
 * @param data the data to store
 * @param size the size of the data being stored
 */
void set_value_at_key(hashmap  *map, char* str, void* data, size_t size);

#endif // hashmap_H