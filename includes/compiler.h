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

void append_to_file(compiler *self, char *str);

void emit_function(compiler *self, function_ast_node *fan);

/**
 * Creates an instance of the Compiler
 * @return the compiler instance
 */
compiler *create_compiler();

/**
 * Starts the compiler
 * @param compiler the compiler instance
 * @param ast the AST to compile
 */
void start_compiler(compiler *compiler, vector *ast);

/**
 * Destroys the given compiler
 * @param compiler the compiler instance to destroy
 */
void destroy_compiler(compiler *compiler);

#endif // compiler_H
