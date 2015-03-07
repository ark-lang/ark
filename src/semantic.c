#include "semantic.h"

semantic* create_semantic_analyser(vector *tree) {
    semantic *self = safe_malloc(sizeof(*self));
    self->tree = tree;
    self->current_node = 0;
    return self;
}

void eat_tasty_node(semantic *self) {
    self->current_node += 1;
}

void start_analysis(semantic *self) {
    while (self->current_node < self->tree->size) {
        printf("yum\n");
        eat_tasty_node(self);
    }
}

void destroy_semantic_analyser(semantic *self) {
    free(self);
}
