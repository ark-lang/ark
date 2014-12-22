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
 
hashmap *create_hashmap(int);

void destroy_hashmap(hashmap *map);

void *get_value_at_key(hashmap *map, char* str);

void set_value_at_key(hashmap  *map, char* str, void* data, size_t size);

#endif // hashmap_H