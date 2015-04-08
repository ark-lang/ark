#ifndef __SEMANTIC_H
#define __SEMANTIC_H

#include "vector.h"
#include "hashmap.h"

typedef struct {
	Vector *parseTree;
	map_t symTable;
} SemanticAnalysis;

SemanticAnalysis *createSemanticAnalysis();

void destroySemanticAnalysis(SemanticAnalysis *self);

#endif // __SEMANTIC_H