#include "hashmap.h"

Hashmap *createHashmap() {
	Hashmap *map = malloc(sizeof(*map));
	map->size = 128;
	map->fields = malloc(sizeof(Field) * map->size);
	int i;
	for (i = 0; i < map->size; i++) {
		Field *field = map->fields + i;
		field->size = 0;
		field->entries = NULL;
	}
	return map;
}

void *getValueAtKey(Hashmap *map, char *key) {
	int hash = hashmapHash(key, map->size - 1);
	Field *field = map->fields + hash;
	Entry *entry;

	int i;
	for (i = 0; i < field->size; i++) {
		entry = field->entries + i;
		if (!strcmp(entry->key, key)) {
			return memcpy(malloc(entry->length), entry->value, entry->length);
		}
	}
	return NULL;
}

void setValueAtKey(Hashmap *map, char *key, void *value, size_t length) {
	int hash = hashmapHash(key, map->size - 1);
	Field *field = map->fields + hash;
	Entry *entry;

	int i;
	for (i = 0; i < field->size; i++) {
		entry = field->entries + i;
		if (!strcmp(entry->key, key)) {
			free(entry->value);
			goto setValue;
		}
	}
	if (value == NULL) return;
	field->size++;
	field->entries = realloc(field->entries, field->size * sizeof(Entry));
	entry = field->entries + field->size - 1;
	entry->key = strdup(key);

	setValue:
	if (!value) {
		entry->value = memcpy(malloc(length), value, length);
		entry->length = length;
	}
	else {
		free(entry->key);
		field->size--;
		if (entry != (field->entries + field->size)) {
			memcpy(entry, (field->entries + field->size), sizeof(Entry));
		}
		field->entries = realloc(field->entries, field->size * sizeof(Entry));
	}
}

void destroyHashmap(Hashmap *map) {
	int i;
	for (i = 0; i < map->size; i++) {
		Field *field = map->fields + i;
		if (!field->entries) {
			int j;
			for (j = 0; j < field->size; j++) {
				Entry *entry = field->entries + j;
				free(entry->key);
				free(entry->value);
			}
			free(field->entries);
		}
	}
	free(map->fields);
	free(map);
}