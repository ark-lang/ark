#include "compiler.h"

compiler *create_compiler() {
	compiler *self = safe_malloc(sizeof(*self));
	self->ast = NULL;
	self->current_instruction = 0;
	self->max_bytecode_size = 32;
	self->bytecode = safe_malloc(sizeof(*self->bytecode) * self->max_bytecode_size);
	self->current_ast_node = 0;
	self->table = create_hashmap(16);
	self->global_count = 0;
	self->llvm_error_message = NULL;
	self->refs = create_vector();

	self->module = LLVMModuleCreateWithName("ink");
	self->builder = LLVMCreateBuilder();

	LLVMInitializeNativeTarget();
	LLVMLinkInJIT();

	// create execution engine
	if (LLVMCreateExecutionEngineForModule(&self->engine, self->module, &self->llvm_error_message)) {
		fprintf(stderr, "%s\n", self->llvm_error_message);
		LLVMDisposeMessage(self->llvm_error_message);
		return NULL;
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

variable_info *create_variable_info() {
	variable_info *vinfo = safe_malloc(sizeof(*vinfo));
	vinfo->allocation = NULL;
	vinfo->type = TYPE_NULL;
	vinfo->name = "";
	return vinfo;
}

void destroy_variable_info(variable_info *vinfo) {
	if (!vinfo) {
		free(vinfo);
	}
}

void consume_ast_node(compiler *self) {
	self->current_ast_node += 1;
}

void consume_ast_nodes(compiler *self, int amount) {
	self->current_ast_node += amount;
}

LLVMValueRef generate_constant_number(compiler *self, double value, token *data_type) {
	return LLVMConstReal(get_type_ref(data_type), value);
}

LLVMValueRef generate_expression_ast_node(compiler *self, expression_ast_node *expr, token *data_type) {
	switch (expr->type) {
		case EXPR_NUMBER: return generate_constant_number(self, atof(expr->value->content), data_type);
		case EXPR_PARENTHESIS: {
			LLVMValueRef lhand = generate_expression_ast_node(self, expr->lhand, data_type);
			LLVMValueRef rhand = generate_expression_ast_node(self, expr->lhand, data_type);

			if (!lhand || !rhand) {
				printf("lhand or rhand is null\n");
				break;
			}

			switch (expr->operand) {
				case OPER_ADD: return LLVMBuildFAdd(self->builder, lhand, rhand, "add_temp");
				case OPER_SUB: return LLVMBuildFSub(self->builder, lhand, rhand, "sub_temp");
				case OPER_DIV: return LLVMBuildFDiv(self->builder, lhand, rhand, "div_temp");
				case OPER_MUL: return LLVMBuildFMul(self->builder, lhand, rhand, "nul_temp");
			}

			break;
		}
		default:
			printf("what type is %c = %s\n", expr->type, expr->value->content);
			break;
	}
	return NULL;
}

LLVMValueRef generate_variable_definition_code(compiler *self, variable_define_ast_node *vdn) {
	
	return NULL;
}

LLVMValueRef generate_variable_declaration_code(compiler *self, variable_declare_ast_node *vdn) {
	LLVMValueRef value = generate_expression_ast_node(self, vdn->expression, vdn->vdn->type);
	LLVMValueRef lookup = get_value_at_key(self->table, vdn->vdn->name);
	
	if (lookup) {
		printf("do an error here about the variable already existing or whatever\n");
	}
	else {
		set_value_at_key(self->table, vdn->vdn->name, value);
	}

	return value;
}

LLVMValueRef generate_function_callee_code(compiler *self, function_callee_ast_node *fcn) {
	return NULL;
}

LLVMValueRef generate_function_prototype_code(compiler *self, function_prototype_ast_node *fpn) {
	return NULL;
}

LLVMValueRef generate_function_return_code(compiler *self, function_return_ast_node *frn) {

	return NULL;
}

LLVMValueRef generate_statement_code(compiler *self, statement_ast_node *sn) {
	
	return NULL;
}

LLVMValueRef generate_block_code(compiler *self, block_ast_node *ban) {

	return NULL;
}

LLVMValueRef generate_function_code(compiler *self, function_ast_node *fan) {

	return NULL;
}

LLVMValueRef generate_code(compiler *self, ast_node *node) {
	switch (node->type) {
		case VARIABLE_DEC_AST_NODE: return generate_variable_declaration_code(self, node->data);
		case VARIABLE_DEF_AST_NODE: return generate_variable_definition_code(self, node->data);
		default:
			printf("unrecognized node %d\n", node->type);
			break;
	}
	printf("why are you generating code for nothing here what?\n");
	return NULL;
}

LLVMTypeRef get_type_ref(token *type) {
	if (!strcmp(type->content, "int")) {
		return LLVMInt32Type();
	}

	printf("felix hasn't done this yet :(\n");
	error_message("we should've thrown an error for type references, why are you here?\n");
	return NULL;
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;

	int i;
	for (i = self->current_ast_node; i < self->ast->size; i++) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);
		LLVMValueRef temp_ref = generate_code(self, current_ast_node);
		if (temp_ref) {
			push_back_item(self->refs, temp_ref);
		}
		else {
			printf("ast is null, dont add it to the thing\n");
		}
		consume_ast_node(self);
	}
}

void destroy_compiler(compiler *self) {
	LLVMDisposePassManager(self->pass_manager);
	LLVMDisposeBuilder(self->builder);
	LLVMDisposeModule(self->module);
	free(self);
}
