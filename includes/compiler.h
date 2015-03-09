#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "hashmap.h"

typedef struct {
	vector *ast;
	vector *refs;
	hashmap *table;

	int *bytecode;
	int global_count;
	int initial_bytecode_size;
	int max_bytecode_size;
	int current_ast_node;
	int current_instruction;
} compiler;

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
