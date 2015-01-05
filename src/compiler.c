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
