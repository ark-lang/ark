#ifndef SEMANTIC_H
#define SEMANTIC_H

#include "util.h"
#include "vector.h"

typedef struct {
    Vector *tree;
    int current_node;
} SemanticAnalyser;

SemanticAnalyser *createSemanticAnalyser(Vector *tree);

void eatAstNode(SemanticAnalyser *self);

void startSemanticAnalysis(SemanticAnalyser *self);

void destroySemanticAnalyser(SemanticAnalyser *self);

#endif // SEMANTIC_H
