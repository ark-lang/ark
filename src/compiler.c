#include "compiler.h"

compiler *create_compiler() {
	compiler *self = safe_malloc(sizeof(*self));
	self->ast = NULL;
	self->current_ast_node = 0;

	self->file_size = 128;
	self->file_cursor_pos = 0;
	self->file_name = "test.c";
	self->file_contents = malloc(sizeof(char) * (self->file_size + 1));
	self->table = create_hashmap(16);

	return self;
}

void append_to_file(compiler *self, char *str) {
	size_t len = strlen(str);
	if (self->file_cursor_pos + len >= self->file_size) {
		self->file_size *= 2;
		self->file_contents = realloc(self->file_contents, sizeof(char) * (self->file_size + 1));
		
		if (!self->file_contents) {
			perror("realloc: failed to re-allocate memory for file contents");
			// exit or whatever
		}
	}

	// pls work
	self->file_contents = strcat(self->file_contents, str);
	self->file_cursor_pos += len;
	self->file_contents[self->file_cursor_pos] = '\0'; 
}

void emit_block(compiler *self, block_ast_node *ban) {
	switch (ban->type) {
		
	}
}

void emit_function(compiler *self, function_ast_node *fan) {
	char *return_type = fan->fpn->return_type->content;

	append_to_file(self, return_type);
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, fan->fpn->name->content);
	append_to_file(self, OPEN_BRACKET);

	int i;
	for (i = 0; i < fan->fpn->args->size; i++) {
		function_argument_ast_node *current = get_vector_item(fan->fpn->args, i);
		if (current->is_constant) {
			append_to_file(self, CONST_KEYWORD);
			append_to_file(self, SPACE_CHAR);
		}

		append_to_file(self, current->type->content);
		append_to_file(self, SPACE_CHAR);
		if (current->is_pointer) append_to_file(self, ASTERISKS);
		append_to_file(self, current->name->content);
	}

	append_to_file(self, CLOSE_BRACKET);
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, OPEN_BRACE);

	printf("%s\n", self->file_contents);
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
		switch (current_ast_node->type) {
			case FUNCTION_AST_NODE: emit_function(self, current_ast_node->data); break;
			default:
				printf("wat ast node bby?\n");
				break;
		}
		consume_ast_node(self);
	}
}

void destroy_compiler(compiler *self) {
	// destroy table
	free(self);
}
