#include "hashmap.h"
 
unsigned long int fnv1a(void *data, unsigned long int len) {
	unsigned char *p = (unsigned char *) data;
	unsigned long int h = 2166136261UL;
	unsigned long int i;
 
	for(i = 0; i < len; i++) {
		h = (h ^ p[i]) * 16777619;
	}
 
	return h;
}
 
/* Maps the hash to the interval [0, max_hash] */
int hashString(char *str, int max_hash) {
	unsigned long int fnv1a_hash = fnv1a((void*) str, strlen(str));
	int hash = (int) (fnv1a_hash % (max_hash + 1));
	return hash;
}
 
/* Creates a new hashmap */
Hashmap *createHashmap(int size) {
	Hashmap *map = malloc(sizeof(Hashmap));
	map->size = size;
	map->fields = malloc(sizeof(HashmapField) * size);
	int i;
	for(i = 0; i < size; i++) {
		HashmapField *field = map->fields + i;
		field->size = 0;
		field->entries = NULL;
	}
	return map;
}
 
/* Frees the hashmap. You have to manually set the pointer to NULL after using this function */
void destroyHashmap(Hashmap *map) {
	int i;
	for(i = 0; i < map->size; i++) {
		HashmapField *field = map->fields + i;
		if(field->entries != 0) {
			int j;
			for(j = 0; j < field->size; j++) {
				HashmapEntry *entry = field->entries + j;
				free(entry->key);
				free(entry->val);
			}
			free(field->entries);
		}
	}
	free(map->fields);
	free(map);
}
 
/* Sets a value in the hashmap. The memory area that val points to is copied. */
void setValueAtKey(Hashmap *map, char *key, void *value, size_t length) {
	int hash = hashString(key, map->size - 1);
	HashmapField *field = map->fields + hash;
	HashmapEntry *entry;
 
	int i;
	/* Check if entry with the same key already exists in field. */
	for(i = 0; i < field->size; i++) {
		entry = field->entries + i;
		if(strcmp(entry->key, key) == 0) {
			/* Depending on value entry is deleted
			or another value is set. In both cases the
			old val can be freed */
			free(entry->val);
			goto set_val;
		}
	}
	/* Create new entry */
	if(value == NULL) return;
	field->size++;
	field->entries = realloc(field->entries, field->size * sizeof(HashmapEntry));
	entry = field->entries + field->size - 1;
	entry->key = strdup(key);
 
	set_val:
	if(value != NULL) {
		entry->val = memcpy(malloc(length), value, length);
		entry->len = length;
	} else {
		/* val is already freed. Key is left */
		free(entry->key);
		field->size--;
		/* Copy last entry to new position */
		if(entry != (field->entries + field->size)) {
			memcpy((void*) entry, (void*) (field->entries + field->size), sizeof(HashmapEntry));
		}
		/* Shrink field */
		field->entries = realloc((void*) field->entries, field->size * sizeof(HashmapEntry));
	}
}
 
/* Get a void-ptr from the hash-map. It is copied from the hashmap and must be freed manually. */
void *getValueAtKey(Hashmap *map, char *key) {
	int hash = hashString(key, map->size - 1);
	HashmapField *field = map->fields + hash;
	HashmapEntry *entry;
 
	int i;
	for(i = 0; i < field->size; i++) {
		entry = field->entries + i;
		if(strcmp(entry->key, key) == 0) {
			return memcpy(malloc(entry->len), entry->val, entry->len);
		}
	}
	return NULL;
}