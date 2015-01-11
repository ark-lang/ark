#include "compiler.h"

compiler *create_compiler() {
	compiler *self = safe_malloc(sizeof(*self));
	self->ast = NULL;
	self->current_instruction = 0;
	self->max_bytecode_size = 32;
	self->bytecode = safe_malloc(sizeof(*self->bytecode) * self->max_bytecode_size);
	self->current_ast_node = 0;
	self->table = create_hashmap(128);
	self->global_count = 0;
	self->llvm_error_message = NULL;
	self->refs = create_vector();

	self->module = LLVMModuleCreateWithName("j4");
	self->builder = LLVMCreateBuilder();

	LLVMInitializeNativeTarget();
	LLVMLinkInJIT();

	// create execution engine
	if (LLVMCreateExecutionEngineForModule(&self->engine, self->module, &self->llvm_error_message)) {
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

LLVMTypeRef get_type_ref(data_type type) {
	printf("given %d\n", type);
	switch (type) {
		case TYPE_INTEGER: 
			printf("integer\n");
			return LLVMInt32Type();
		case TYPE_STR:
			error_message("strings are unimplemented\n");
			break;
		case TYPE_DOUBLE: printf("double\n"); return LLVMDoubleType();
		case TYPE_FLOAT: printf("Float\n"); return LLVMFloatType();
		case TYPE_BOOL:
			error_message("bools are unimplemented\n");
			break;
		case TYPE_VOID: printf("Void\n"); return LLVMVoidType();
		case TYPE_CHAR:
			error_message("chars are unimplemented\n");
			break;
		default:
			error_message("unsupported data type `%d`\n", type);
			break;
	}
	error_message("we should've thrown an error for type references, why are you here?\n");
	return NULL;
}

LLVMValueRef evaluate_expression_ast_node(compiler *self, expression_ast_node *expr) {
	error_message("expressions unimplemented\n");
	return NULL;
}

LLVMValueRef generate_variable_definition_code(compiler *self, variable_declare_ast_node *vdn) {
	error_message("definitions unimplemented\n");
	return NULL;
}

LLVMValueRef generate_variable_declaration_code(compiler *self, variable_declare_ast_node *vdn) {
	error_message("variable declaration unimplemented\n");
	return NULL;
}

LLVMValueRef generate_function_callee_code(compiler *self, function_callee_ast_node *fcn) {
	LLVMValueRef func = LLVMGetNamedFunction(self->module, fcn->callee->content);
	if (!func) {
		printf("function `%s` not found in module!\n", fcn->callee->content);
	}

	if (LLVMCountParams(func) != fcn->args->size) {
		printf("number of arguments given doesn't match required argument size\n");
	}

	LLVMValueRef *args = safe_malloc(sizeof(LLVMValueRef) * fcn->args->size);
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

LLVMValueRef generate_function_prototype_code(compiler *self, function_prototype_ast_node *fpn) {
	unsigned int i;
	unsigned int arg_count = fpn->args->size;

	LLVMValueRef proto = LLVMGetNamedFunction(self->module, fpn->name->content);
	if (proto) {
		if (LLVMCountParams(proto) != arg_count) {
			error_message("function `%s` with different argument count already exists\n", fpn->name->content);
			return NULL;
		}

		if (LLVMCountBasicBlocks(proto)) {
			error_message("function `%s` exists with a body\n", fpn->name->content);
			return NULL;
		}

		error_message("idk some shit up with the proto for `%s`\n", fpn->name->content);
	}
	else {
		LLVMTypeRef *params = NULL;
		if (!arg_count) {
			params = safe_malloc(sizeof(LLVMTypeRef) * arg_count);

			for (i = 0; i < arg_count; i++) {
				function_argument_ast_node *arg = get_vector_item(fpn->args, i);
				params[i] = get_type_ref(arg->type);
			}
		}

		// get the first argument for now, tuples aren't supported just yet
		data_type *return_val = get_vector_item(fpn->ret, 0);
		

		LLVMTypeRef return_type = get_type_ref(*return_val);
		LLVMTypeRef func_type = LLVMFunctionType(return_type, params, arg_count, false);

		proto = LLVMAddFunction(self->module, fpn->name->content, func_type);
	}

	for (i = 0; i < arg_count; i++) {
		function_argument_ast_node *arg = get_vector_item(fpn->args, i);
		LLVMValueRef param = LLVMGetParam(proto, i);
		LLVMSetValueName(param, arg->name->content);

		set_value_at_key(self->table, arg->name->content, arg, sizeof(arg));
	}

	return proto;
}

LLVMValueRef generate_function_return_code(compiler *self, function_return_ast_node *frn) {

	return NULL;
}

LLVMValueRef generate_statement_code(compiler *self, statement_ast_node *sn) {
	switch (sn->type) {
		case VARIABLE_DEF_AST_NODE:
			return generate_variable_definition_code(self, sn->data);
		case VARIABLE_DEC_AST_NODE:
			return generate_variable_declaration_code(self, sn->data);
		case FUNCTION_CALLEE_AST_NODE:
			return generate_function_callee_code(self, sn->data);
		case FUNCTION_RET_AST_NODE:
			return generate_function_return_code(self, sn->data);
		case VARIABLE_REASSIGN_AST_NODE:
			primary_message("variable reassign unimplemented\n");
			break;
		case FOR_LOOP_AST_NODE:
			primary_message("for loop unimplemented\n");
			break;
		case INFINITE_LOOP_AST_NODE:
			primary_message("infinite loop unimplemented\n");
			break;
		case BREAK_AST_NODE:
			primary_message("break unimplemented\n");
			break;
		case ENUM_AST_NODE:
			primary_message("enum unimplemented\n");
			break;
		default: 
			error_message("unknown node given %d\n", sn->type);
			break;
	}

	printf("why is it returning null?\n");
	return NULL;
}

LLVMValueRef generate_block_code(compiler *self, block_ast_node *ban) {
	int i;
	for (i = 0; i < ban->statements->size; i++) {
		statement_ast_node *sn = get_vector_item(ban->statements, i);
		LLVMValueRef location = NULL;
		LLVMValueRef statement_code = generate_statement_code(self, sn);
		LLVMBuildStore(self->builder, statement_code, location);
	}
	return NULL; // temporary
}

LLVMValueRef generate_function_code(compiler *self, function_ast_node *fan) {
	// first we generate the prototype
	LLVMValueRef func = generate_function_prototype_code(self, fan->fpn);
	if (!func) {
		error_message("prototype for `%s` is dun goofed\n", fan->fpn->name->content);
		return NULL;
	}

	LLVMBasicBlockRef block = LLVMAppendBasicBlock(func, "entry");
	LLVMPositionBuilderAtEnd(self->builder, block);

	LLVMValueRef body = generate_block_code(self, fan->body);
	if (!body) {
		error_message("failed to generate code for `%s`'s body\n", fan->fpn->name->content);
	}

	LLVMBuildRet(self->builder, body);

	if (LLVMVerifyFunction(func, LLVMPrintMessageAction)) {
		error_message("invalid function `%s`\n", fan->fpn->name->content);
		LLVMDeleteFunction(func);
		return NULL;
	}

	return func;
}

LLVMValueRef generate_code(compiler *self, ast_node *node) {
	switch (node->type) {
		case VARIABLE_DEC_AST_NODE:
			return generate_variable_declaration_code(self, node->data);
		case FUNCTION_AST_NODE:
			return generate_function_code(self, node->data);
		case FUNCTION_CALLEE_AST_NODE:
			return generate_function_callee_code(self, node->data);
		default:
			debug_message("unrecognized node specified", true);
			break;
	}
	debug_message("unknown node, why are you here?");
	return NULL;
}

void start_compiler(compiler *self, vector *ast) {
	self->ast = ast;


	int i;
	for (i = self->current_ast_node; i < self->ast->size; i++) {
		ast_node *current_ast_node = get_vector_item(self->ast, self->current_ast_node);

		LLVMValueRef temp_ref = generate_code(self, current_ast_node);
		if (temp_ref) {
			printf("ast is null, dont add it to the thing\n");
		}
		else {
			push_back_item(self->refs, temp_ref);
		}

		printf("generating code for node at index %d\n", i);
		consume_ast_node(self);
	}
}

void destroy_compiler(compiler *self) {
	if (self != NULL) {
		free(self);
	}
	LLVMDisposePassManager(self->pass_manager);
	LLVMDisposeBuilder(self->builder);
	LLVMDisposeModule(self->module);
}
