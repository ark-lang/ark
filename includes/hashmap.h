#ifndef HASHMAP_H
#define HASHMAP_H

#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "util.h"

typedef struct {
	char *key;
	char *value;
	size_t length;
} Entry;

typedef struct {
	int size;
	Entry *entries;
} Field;

typedef struct {
	int size;
	Field *fields;
} Hashmap;

Hashmap *hashmapCreate();

void hashmapSet(Hashmap *map, char *key, void *value, size_t length);

void *hashmapGet(Hashmap *map, char *key);

void hashmapDestroy(Hashmap *map);

static inline unsigned long int hashmapFNV1A(void *data, unsigned long int length) {
	unsigned char *p = (unsigned char *) data;
	unsigned long int h = 2166136261UL;
	unsigned long int i;
	for (i = 0; i < length; i++) {
		h = (h ^ p[i]) * 16777619;
	}
	return h;
}

static inline int hashmapHash(char *str, int maxHash) {
	unsigned long int fnv1aHash = hashmapFNV1A(str, strlen(str));
	return (int) (fnv1aHash % (maxHash + 1));
}

#endif // HASHMAP_H