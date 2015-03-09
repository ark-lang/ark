#include "compiler.h"

compiler *create_compiler() {
	compiler *self = safe_malloc(sizeof(*self));
	self->ast = NULL;
	self->current_ast_node = 0;

	self->table = create_hashmap(16);

	return self;
}

void consume_ast_node(compiler *self) {
	self->current_ast_node += 1;
}

void consume_ast_nodes(compiler *self, int amount) {
	self->current_ast_node += amount;
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	int i;
	for (i = self->current_ast_node; i < self->ast->size; i++) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);
		// do work
		consume_ast_node(self);
	}
}

void destroy_compiler(compiler *self) {
	// destroy table
	free(self);
}
