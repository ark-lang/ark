#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <llvm-c-3.5/llvm-c/Core.h>
#include <llvm-c-3.5/llvm-c/Analysis.h>
#include <llvm-c-3.5/llvm-c/ExecutionEngine.h>
#include <llvm-c-3.5/llvm-c/Target.h>
#include <llvm-c-3.5/llvm-c/Transforms/Scalar.h>

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
	vector *refs;
	hashmap *table;

	LLVMModuleRef module;
	LLVMBuilderRef builder;
	LLVMExecutionEngineRef engine;
	LLVMPassManagerRef pass_manager;
	char *llvm_error_message;

	int *bytecode;
	int global_count;
	int initial_bytecode_size;
	int max_bytecode_size;
	int current_ast_node;
	int current_instruction;
} compiler;

typedef struct {
	LLVMValueRef allocation;
	char *name;
	data_type_w *type;
} variable_info;

variable_info *create_variable_info();

void destroy_variable_info(variable_info *vinfo);

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
 * Consumes the current node, and skips to the next
 * @param self the compiler instance
 */
void consume_ast_node(compiler *self);

LLVMValueRef generate_code(compiler *self, ast_node *node);

/**
 * Generates code for a function
 * @param self the compiler instance
 * @param func the function node to generate code for
 */
LLVMValueRef generate_function_code(compiler *self, function_ast_node *func);

/**
 * Evaluates an expression node
 * @param self the compiler instance
 * @param expr the expression to generate code for
 */
LLVMValueRef evaluate_expression_ast_node(compiler *self, expression_ast_node *expr);

/**
 * Geneartes code for a Variable Declaration
 * @param compiler the compiler instance
 * @param vdn the Variable Node to generate code for
 */
LLVMValueRef generate_variable_declaration_code(compiler *compiler, variable_declare_ast_node *vdn);

/**
 * Generates code for a function call
 * @param self the compiler instance
 * @param fcn the function callee node to generate code for 
 */
LLVMValueRef generate_function_callee_code(compiler *self, function_callee_ast_node *fcn);

/**
 * Generates code for a function return
 * @param self the compiler instance
 * @param frn the function return node to generate code for 
 */
LLVMValueRef generate_function_return_code(compiler *self, function_return_ast_node *frn);

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
