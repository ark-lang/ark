#include "compiler.h"

compiler *create_compiler() {
	compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	self->current_instruction = 0;
	self->max_bytecode_size = 32;
	self->bytecode = malloc(sizeof(*self->bytecode) * self->max_bytecode_size);
	self->current_ast_node = 0;
	self->functions = create_hashmap(128);
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

int evaluate_expression_ast_node(compiler *self, expression_ast_node *expr) {
	if (expr->value != NULL) {
		append_instruction(self, ICONST);
		append_instruction(self, atoi(expr->value->content));
	}
	else {
		evaluate_expression_ast_node(self, expr->lhand);
		evaluate_expression_ast_node(self, expr->rhand);

		switch (expr->operand) {
			case '+': append_instruction(self, ADD); break;
			case '-': append_instruction(self, SUB); break;
			case '*': append_instruction(self, MUL); break;
			case '/': append_instruction(self, DIV); break;
			case '%': append_instruction(self, MOD); break;
			case '^': append_instruction(self, POW); break;
		}
	}
	return 1337;
}

void generate_variable_declaration_code(compiler *self, variable_declare_ast_node *vdn) {
}

void generateFunctionCalleeCode(compiler *self, function_callee_ast_node *fcn) {
	char *name = fcn->callee->content;
	int *address = get_value_at_key(self->functions, name);
	int numOfArgs = fcn->args->size;

	int i;
	for (i = 0; i < numOfArgs; i++) {
		function_argument_ast_node *fan = get_vector_item(fcn->args, i);
		evaluate_expression_ast_node(self, fan->value);
	}

	append_instruction(self, CALL);
	append_instruction(self, *address);
	append_instruction(self, numOfArgs);
}

void generateFunctionReturnCode(compiler *self, function_return_ast_node *frn) {
	if (frn->numOfReturnValues > 1) {
		printf("tuples not yet supported.\n");
		exit(1);
	}
	// no tuple support, just use first return value for now.
	expression_ast_node *expr = get_vector_item(frn->returnVals, 0);
	evaluate_expression_ast_node(self, expr);
}

void generate_function_code(compiler *self, function_ast_node *func) {
	int address = self->current_instruction;
	set_value_at_key(self->functions, func->fpn->name->content, &address, sizeof(int));
		
	vector *statements = func->body->statements;

	// return stuff
	int i;
	for (i = 0; i < statements->size; i++) {
		statement_ast_node *sn = get_vector_item(statements, i);
		switch (sn->type) {
			case FUNCTION_RET_ast_node:
				generateFunctionReturnCode(self, sn->data);
				break;
			default:
				printf("WHAT ast_nodeS YA GIVIN ME SON?\n");
				break;
		}
	}

	append_instruction(self, RET);
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	while (self->current_ast_node < self->ast->size) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);

		switch (current_ast_node->type) {
		case VARIABLE_DEC_ast_node:
			generate_variable_declaration_code(self, current_ast_node->data);
			break;
		case FUNCTION_ast_node:
			generate_function_code(self, current_ast_node->data);
			break;
		case FUNCTION_CALLEE_ast_node:
			generateFunctionCalleeCode(self, current_ast_node->data);
			break;
		default:
			printf("unrecognized ast_node\n");
			break;
		}

		consume_ast_node(self);
	}

	// stop
	append_instruction(self, HALT);
}

void destroy_compiler(compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
