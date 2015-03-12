#include "compiler.h"

compiler *create_compiler() {
	compiler *self = safe_malloc(sizeof(*self));
	self->ast = NULL;
	self->current_ast_node = 0;

	self->file_size = 128;
	self->file_cursor_pos = 0;
	self->file_name = "test.c";
	self->file_contents = malloc(sizeof(char) * (self->file_size + 1));
	self->file_contents[0] = '\0';
	self->table = create_hashmap(16);

	return self;
}

void append_to_file(compiler *self, char *str) {
	self->file_contents = realloc(self->file_contents, strlen(self->file_contents) + strlen(str) + 1);
	strcat(self->file_contents, str);
}

void emit_expression(compiler *self, expression_ast_node *expr) {
	int i;
	for (i = 0; i < expr->expression_values->size; i++) {
		token *tok = get_vector_item(expr->expression_values, i);
		append_to_file(self, tok->content);
		
		if (i != expr->expression_values->size - 1) {
			append_to_file(self, SPACE_CHAR);
		}
	}
}

void emit_variable_dec(compiler *self, variable_declare_ast_node *var) {
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
	append_to_file(self, call->callee);
	append_to_file(self, OPEN_BRACKET);
	
	// print args	
	emit_arguments(self, call->args);

	append_to_file(self, CLOSE_BRACKET);
	append_to_file(self, SEMICOLON);
}

void emit_if_statement(compiler *self, if_statement_ast_node *stmt) {
	append_to_file(self, "if");
	append_to_file(self, SPACE_CHAR);

	append_to_file(self, OPEN_BRACKET);
	emit_expression(self, stmt->condition);
	append_to_file(self, CLOSE_BRACKET);

	emit_block(self, stmt->body);
	if (stmt->else_statement) {
		append_to_file(self, "else");
		emit_block(self, stmt->else_statement);
	}
}

void emit_block(compiler *self, block_ast_node *block) {
	append_to_file(self, SPACE_CHAR);
	append_to_file(self, OPEN_BRACE);
	append_to_file(self, NEWLINE);

	int i;
	for (i = 0; i < block->statements->size; i++) {
		statement_ast_node *current = get_vector_item(block->statements, i);
		append_to_file(self, TAB);
		switch (current->type) {
			case VARIABLE_DEC_AST_NODE:
				emit_variable_dec(self, current->data);
				break;
			case FUNCTION_CALLEE_AST_NODE:
				emit_function_call(self, current->data);
				break;
			case FUNCTION_RET_AST_NODE:
				emit_return(self, current->data);
				break;
			case IF_STATEMENT_AST_NODE:
				emit_if_statement(self, current->data);
				break;
			default:
				printf("idk fuk off\n");
				break;
		}
		if (i != block->statements->size - 1) {
			append_to_file(self, NEWLINE);
		}
	}

	append_to_file(self, NEWLINE);
	append_to_file(self, CLOSE_BRACE);
	append_to_file(self, NEWLINE);
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

void emit_return(compiler *self, function_return_ast_node *ret) {
	append_to_file(self, "return");
	if (ret->return_val) {
		append_to_file(self, SPACE_CHAR);
		emit_expression(self, ret->return_val);
	}
	append_to_file(self, SEMICOLON);
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

	emit_block(self, func->body);
}

void consume_ast_node(compiler *self) {
	self->current_ast_node += 1;
}

void consume_ast_nodes(compiler *self, int amount) {
	self->current_ast_node += amount;
}

void write_file(compiler *self) {
	FILE *file = fopen("temp.c", "w");
	if (!file) {
		perror("fopen: failed to open file");
		return;
	}

	fprintf(file, "%s", self->file_contents);
	fclose(file);

	system("gcc temp.c");

	// remove the file since we don't need it
	remove("temp.c");
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	// include for standard library
	// eventually I'll write an inclusion system
	// and some way of calling C functions
	append_to_file(self, "#include <stdio.h>" NEWLINE);

	int i;
	for (i = 0; i < self->ast->size; i++) {
		ast_node *current_ast_node = get_vector_item(self->ast, i);

		switch (current_ast_node->type) {
			case FUNCTION_AST_NODE: 
				emit_function(self, current_ast_node->data); 
				break;
			default:
				printf("wat ast node bby %d is %d?\n", current_ast_node->type, i);
				break;
		}
	}

	printf("%s\n", self->file_contents);
	write_file(self);
}

void destroy_compiler(compiler *self) {
	// destroy table
	free(self);
}
