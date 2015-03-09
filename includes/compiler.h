#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "hashmap.h"

/**
 * For now we just assume that the code is semantically correct,
 * we should really semantically analyze everything though, but thats
 * not as fun as code generation :)
 *
 * inb4 vedant trys to remove this because hes a shithead
 */

typedef struct {
	vector *ast;
	vector *refs;
	hashmap *table;

	char *file_name;
	char *file_contents;

	int current_ast_node;
	int current_instruction;
} compiler;

void emit_function(char *function_name, block_ast_node *block);

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
