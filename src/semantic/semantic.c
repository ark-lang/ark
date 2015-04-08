#include "semantic.h"

SemanticAnalysis *createSemanticAnalysis(Vector *parseTree) {
	SemanticAnalysis *self = safeMalloc(sizeof(*self));
	self->symTable = hashmap_new();
	self->parseTree = parseTree;
	return self;
}

void destroySemanticAnalysis(SemanticAnalysis *self) {
	hashmap_free(self->symTable);
	free(self);
}
