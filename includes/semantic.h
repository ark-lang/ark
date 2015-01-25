#ifndef SEMANTIC_H
#define SEMANTIC_H

#include "util.h"
#include "vector.h"

typedef struct {
    vector *tree;
    int current_node;
} semantic;

semantic *create_semantic_analyser(vector *tree);

void eat_tasty_node(semantic *self);

void start_analysis(semantic *self);

void destroy_semantic_analyser(semantic *self);

#endif // SEMANTIC_H