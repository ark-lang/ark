#include "semantic.h"

SemanticAnalysis *createSemanticAnalysis(Vector *parseTree) {
	SemanticAnalysis *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->parseTree = parseTree;
	return self;
}

void analyzeUnstructuredStatement(SemanticAnalysis *self, UnstructuredStatement *unstructured) {
	
}

void analyzeStructuredStatement(SemanticAnalysis *self, StructuredStatement *unstructured) {
	
}

void analyzeStatement(SemanticAnalysis *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			analyzeUnstructuredStatement(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE: 
			analyzeStructuredStatement(self, stmt->structured);
			break;
	}
}

void startSemanticAnalysis(SemanticAnalysis *self) {
	for (int i = 0; i < self->parseTree->size; i++) {
		Statement *stmt = getVectorItem(self->parseTree, i);
		analyzeStatement(self, stmt);
	}
}

void destroySemanticAnalysis(SemanticAnalysis *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	free(self);
}
