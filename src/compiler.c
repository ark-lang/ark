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
	self->llvm_error_message = NULL;
	self->refs = create_vector();

	self->module = LLVMModuleCreateWithName("j4");
	self->builder = LLVMCreateBuilder();
	LLVMInitializeNativeTarget();
	LLVMLinkInJIT();

	// create execution engine
	if (!LLVMCreateExecutionEngineForModule(&self->engine, self->module, &self->llvm_error_message)) {
		fprintf(stderr, "%s\n", self->llvm_error_message);
		LLVMDisposeMessage(self->llvm_error_message);
		exit(1);
	}

	// optimizations
	self->pass_manager = LLVMCreateFunctionPassManagerForModule(self->module);
	LLVMAddTargetData(LLVMGetExecutionEngineTargetData(self->engine), self->pass_manager);
	LLVMAddPromoteMemoryToRegisterPass(self->pass_manager);
	LLVMAddInstructionCombiningPass(self->pass_manager);
	LLVMAddReassociatePass(self->pass_manager);
	LLVMAddGVNPass(self->pass_manager);
	LLVMAddCFGSimplificationPass(self->pass_manager);
	LLVMInitializeFunctionPassManager(self->pass_manager);

	return self;
}

void consume_ast_node(compiler *self) {
	self->current_ast_node += 1;
}

void consume_ast_nodes(compiler *self, int amount) {
	self->current_ast_node += amount;
}

LLVMValueRef evaluate_expression_ast_node(compiler *self, expression_ast_node *expr) {
	return NULL;
}

LLVMValueRef generate_variable_declaration_code(compiler *self, variable_declare_ast_node *vdn) {
	return NULL;
}

LLVMValueRef generate_function_callee_code(compiler *self, function_callee_ast_node *fcn) {
	LLVMValueRef func = LLVMGetNamedFunction(self->module, fcn->callee->content);
	if (!func) {
		printf("function not found in module!\n");
	}

	if (LLVMCountParams(func) != fcn->args->size) {
		printf("number of arguments given doesn't match required argument size\n");
	}

	LLVMValueRef *args = malloc(sizeof(LLVMValueRef) * fcn->args->size);
	unsigned int i;
	unsigned int arg_count = fcn->args->size;
	for (i = 0; i < arg_count; i++) {
		args[i] = NULL;
		if (args[i] == NULL) {
			free(args);
			printf("invalid argument given do some error here\n");
		}
	}
	return LLVMBuildCall(self->builder, func, args, arg_count, "calltmp");
}

LLVMValueRef generate_function_return_code(compiler *self, function_return_ast_node *frn) {
	return NULL;
}

LLVMValueRef generate_function_code(compiler *self, function_ast_node *func) {
	return NULL;
}

LLVMValueRef generate_code(compiler *self, ast_node *node) {
	switch (node->type) {
		case VARIABLE_DEC_AST_NODE:
			return generate_variable_declaration_code(self, node->data);
			break;
		case FUNCTION_AST_NODE:
			return generate_function_code(self, node->data);
			break;
		case FUNCTION_CALLEE_AST_NODE:
			return generate_function_callee_code(self, node->data);
			break;
		case FUNCTION_RET_AST_NODE:
			return generate_function_return_code(self, node->data);
			break;
		default:
			debug_message("unrecognized node specified", true);
			break;
	}
	debug_message("unknown node, why are you here?");
	return NULL;
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	while (self->current_ast_node < self->ast->size) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);

		LLVMValueRef temp_ref = generate_code(self, current_ast_node);
		if (temp_ref) {
			printf("ast is null, dont add it to the thing\n");
		}
		else {
			push_back_item(self->refs, temp_ref);
		}
		consume_ast_node(self);
	}
}

void destroy_compiler(compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
	LLVMDumpModule(self->module);
	LLVMDisposePassManager(self->pass_manager);
	LLVMDisposeBuilder(self->builder);
	LLVMDisposeModule(self->module);
}
