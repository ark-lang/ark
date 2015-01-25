#include "semantic.h"

semantic *create_semantic_analyser(vector *tree) {
    semantic *self = safe_malloc(sizeof(*self));
    self->tree = tree;
    return self;
}

void destroy_semantic_analyser(semantic *self) {
    free(self);
}