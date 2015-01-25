#ifndef SEMANTIC_H
#define SEMANTIC_H

#include "util.h"
#include "vector.h"

typedef struct {
    vector *tree;
} semantic;

semantic *create_semantic_analyser(vector *tree);

void destroy_semantic_analyser(semantic *self);

#endif // SEMANTIC_H