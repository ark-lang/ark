#ifndef HASHMAP_H
#define HASHMAP_H

#include <stdlib.h>
#include <string.h>

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
 
Hashmap *createHashmap(int);

void destroyHashmap(Hashmap *map);

void *getValueAtKey(Hashmap *map, char* str);

void setValueAtKey(Hashmap  *map, char* str, void* data, size_t size);

#endif // HASHMAP_H