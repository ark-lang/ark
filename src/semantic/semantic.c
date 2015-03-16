#include "semantic/semantic.h"

SemanticAnalyser *createSemanticAnalyser(Vector *tree) {
    SemanticAnalyser *self = safeMalloc(sizeof(*self));
    self->tree = tree;
    self->current_node = 0;
    return self;
}

void eatAstNode(SemanticAnalyser *self) {
    self->current_node += 1;
}

void startSemanticAnalysis(SemanticAnalyser *self) {
    while (self->current_node < self->tree->size) {
        eatAstNode(self);
    }
}

void destroySemanticAnalyser(SemanticAnalyser *self) {
    free(self);
}
