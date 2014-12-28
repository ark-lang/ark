#include "compiler.h"

compiler *create_compiler() {
	compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	self->current_instruction = 0;
	self->max_bytecode_size = 32;
	self->bytecode = malloc(sizeof(*self->bytecode) * self->max_bytecode_size);
	self->current_ast_node = 0;
	self->functions = create_hashmap(128);
	self->global_count = 0;
	return self;
}

void append_instruction(compiler *self, int instr) {
	if (self->current_instruction >= self->max_bytecode_size) {
		self->max_bytecode_size *= 2;
		self->bytecode = realloc(self->bytecode, sizeof(*self->bytecode) * self->max_bytecode_size);
	}
	self->bytecode[self->current_instruction++] = instr;
}

void consume_ast_node(compiler *self) {
	self->current_ast_node += 1;
}

void evaluate_expression_ast_node(compiler *self, expression_ast_node *expr) {
	// TODO
}

void generate_variable_declaration_code(compiler *self, variable_declare_ast_node *vdn) {
	// TODO
}

void generateFunctionCalleeCode(compiler *self, function_callee_ast_node *fcn) {
	// TODO
}

void generateFunctionReturnCode(compiler *self, function_return_ast_node *frn) {
	// TODO
}

void generate_function_code(compiler *self, function_ast_node *func) {
	// TODO	
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	while (self->current_ast_node < self->ast->size) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);

		switch (current_ast_node->type) {
		case VARIABLE_DEC_AST_NODE:
			generate_variable_declaration_code(self, current_ast_node->data);
			break;
		case FUNCTION_AST_NODE:
			generate_function_code(self, current_ast_node->data);
			break;
		case FUNCTION_CALLEE_AST_NODE:
			generateFunctionCalleeCode(self, current_ast_node->data);
			break;
		default:
			debug_message("unrecognized node specified", true);
			break;
		}

		consume_ast_node(self);
	}

	append_instruction(self, HALT);
}

void destroy_compiler(compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
