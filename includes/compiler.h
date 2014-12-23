#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "j4vm.h"
#include "hashmap.h"

typedef struct {
	vector *ast;
	jayfor_vm *vm;
	hashmap *functions;
	int *bytecode;

	int initial_bytecode_size;
	int max_bytecode_size;
	int current_ast_node;
	int current_instruction;
} compiler;

compiler *create_compiler();

void append_instruction(compiler *self, int instr);

void generate_function_code(compiler *self, function_ast_node *func);

int evaluate_expression_ast_node(compiler *self, expression_ast_node *expr);

void consume_ast_node(compiler *self);

void generate_variable_declaration_code(compiler *compiler, variable_declare_ast_node *vdn);

void start_compiler(compiler *compiler, vector *ast);

void destroy_compiler(compiler *compiler);

#endif // compiler_H