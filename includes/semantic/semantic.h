#ifndef __SEMANTIC_H
#define __SEMANTIC_H

/**
 * Semantic analysis of our parser, not a high priority as
 * of writing this, so it's partially unimplemented.
 */

#include "ast.h"
#include "vector.h"
#include "hashmap.h"

typedef struct {
	Vector *parseTree;

	map_t funcSymTable;

	map_t varSymTable;
} SemanticAnalysis;

SemanticAnalysis *createSemanticAnalysis();

void analyzeUnstructuredStatement(SemanticAnalysis *self, UnstructuredStatement *unstructured);

void analyzeStructuredStatement(SemanticAnalysis *self, StructuredStatement *unstructured);

void analyzeStatement(SemanticAnalysis *self, Statement *stmt);

void startSemanticAnalysis(SemanticAnalysis *self);

void destroySemanticAnalysis(SemanticAnalysis *self);

#endif // __SEMANTIC_H