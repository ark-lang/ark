#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "hashmap.h"

#define SPACE_CHAR " "
#define OPEN_BRACKET "("
#define CLOSE_BRACKET ")"
#define OPEN_BRACE "{"
#define CLOSE_BRACE "}"
#define CONST_KEYWORD "const"
#define ASTERISKS "*"
#define NEWLINE "\n"
#define TAB "\t"
#define EQUAL_SYM "="
#define SEMICOLON ";"
#define COMMA_SYM ","

typedef struct {
	vector *ast;
	vector *refs;
	hashmap *table;

    int file_size;
    int file_cursor_pos;
	char *file_name;
	char *file_contents;

	int current_ast_node;
	int current_instruction;
} compiler;

compiler *create_compiler();

void append_to_file(compiler *self, char *str);

void emit_expression(compiler *self, expression_ast_node *expr);

void emit_variable_dec(compiler *self, variable_declare_ast_node *var);

void emit_return(compiler *self, function_return_ast_node *ret);

void emit_function_call(compiler *self, function_callee_ast_node *call);

void emit_block(compiler *self, block_ast_node *block);

void emit_arguments(compiler *self, vector *args);

void emit_function(compiler *self, function_ast_node *func);

void consume_ast_node(compiler *self);

void consume_ast_nodes(compiler *self, int amount);

void start_compiler(compiler *self, vector *ast);

void destroy_compiler(compiler *self);

#endif // compiler_H
