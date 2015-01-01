#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <llvm-c/Core.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Transforms/Scalar.h>

#include "parser.h"
#include "vector.h"
#include "hashmap.h"

/**
 * Compiler, takes our AST and
 * generates code for it. Also keeps
 * track of functions, etc. Optimisations
 * are on the todo list, but I want to
 * get it working first.
 */
typedef struct {
	vector *ast;
	hashmap *functions;
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
 * Appends an instruction to our bytecode that we
 * are generating
 * @param self the compiler instance
 * @param instr the instruction to append
 */
void append_instruction(compiler *self, int instr);

/**
 * Generates code for a function
 * @param self the compiler instance
 * @param func the function node to generate code for
 */
void generate_function_code(compiler *self, function_ast_node *func);

/**
 * Evaluates an expression node
 * @param self the compiler instance
 * @param expr the expression to generate code for
 */
void evaluate_expression_ast_node(compiler *self, expression_ast_node *expr);

/**
 * Consumes the current node, and skips to the next
 * @param self the compiler instance
 */
void consume_ast_node(compiler *self);

/**
 * Geneartes code for a Variable Declaration
 * @param compiler the compiler instance
 * @param vdn the Variable Node to generate code for
 */
void generate_variable_declaration_code(compiler *compiler, variable_declare_ast_node *vdn);

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
