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

void emit_expression(compiler *self, expression_ast_node *expr) {
	append_to_file(self, "0"); // THIS IS TEMPORARY TILL I REWORK THE EXPRESSIONS PARSING!!
}

void emit_variable_dec(compiler *self, variable_declare_ast_node *var) {
	append_to_file(self, TAB);
	if (var->vdn->is_constant) {
		append_to_file(self, CONST_KEYWORD);
		append_to_file(self, SPACE_CHAR);
	}
	append_to_file(self, var->vdn->type->content);
	append_to_file(self, SPACE_CHAR);
	if (var->vdn->is_pointer) {
		append_to_file(self, ASTERISKS);
	}
	append_to_file(self, var->vdn->name);
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, EQUAL_SYM);
	append_to_file(self, SPACE_CHAR);
	emit_expression(self, var->expression);
	append_to_file(self, SEMICOLON);
}

void emit_function_call(compiler *self, function_callee_ast_node *call) {
	append_to_file(self, TAB);
	append_to_file(self, call->callee);
	append_to_file(self, OPEN_BRACKET);
	emit_arguments(self, call->args);
	append_to_file(self, CLOSE_BRACKET);
}

void emit_block(compiler *self, block_ast_node *block) {
	int i;
	for (i = 0; i < block->statements->size; i++) {
		statement_ast_node *current = get_vector_item(block->statements, i);
		switch (current->type) {
			case VARIABLE_DEC_AST_NODE:
				emit_variable_dec(self, current->data);
				break;
			case FUNCTION_CALLEE_AST_NODE:
				emit_function_call(self, current->data);
				break;
			default:
				printf("idk fuk off\n");
				break;
		}
		append_to_file(self, NEWLINE);
	}
}

void emit_arguments(compiler *self, vector *args) {
	int i;
	for (i = 0; i < args->size; i++) {
		function_argument_ast_node *current = get_vector_item(args, i);

		if (current->is_constant) {
			append_to_file(self, CONST_KEYWORD);
			append_to_file(self, SPACE_CHAR);
		}

		if (current->type) {
			append_to_file(self, current->type->content);
			append_to_file(self, SPACE_CHAR);
		}

		if (current->is_pointer) append_to_file(self, ASTERISKS);

		// this could fail if we're doing a function call
		if (current->name) {
			append_to_file(self, current->name->content);
		}
		if (current->value) {
			emit_expression(self, current->value);
		}

		if (args->size > 1 && i != args->size - 1) {
			append_to_file(self, COMMA_SYM);
			append_to_file(self, SPACE_CHAR);
		}
	}
}

void emit_function(compiler *self, function_ast_node *func) {
	char *return_type = func->fpn->return_type->content;

	append_to_file(self, return_type);
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, func->fpn->name->content);
	append_to_file(self, OPEN_BRACKET);

	emit_arguments(self, func->fpn->args);	

	append_to_file(self, CLOSE_BRACKET);
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, OPEN_BRACE);

	append_to_file(self, NEWLINE);

	emit_block(self, func->body);

	append_to_file(self, NEWLINE);

	append_to_file(self, CLOSE_BRACE);
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

	printf("%s\n", self->file_contents);
}

void destroy_compiler(compiler *self) {
	// destroy table
	free(self);
}
