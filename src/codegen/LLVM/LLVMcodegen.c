#include "LLVM/LLVMcodegen.h"

#define genError(...) errorMessage("LLVM codegen: " __VA_ARGS__)

// Declarations

/**
 * Jumps ahead in the AST we're parsing
 * @param self   the code gen instance
 * @param amount the amount to consume by
 */
static void consumeAstNodeBy(LLVMCodeGenerator *self, int amount);

/**
 * Run through all the nodes in the AST and
 * generate the code for them!
 * @param self the code gen instance
 */
static void traverseAST(LLVMCodeGenerator *self);

/**
 * Gets the int type for the current system
 */
static LLVMTypeRef getIntType();

static LLVMTypeRef getLLVMType(DataType type);

LLVMValueRef genFunctionSignature(LLVMCodeGenerator *self, FunctionSignature *decl);

LLVMValueRef genStatement(LLVMCodeGenerator *self, Statement *stmt);

LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl);

LLVMValueRef genDeclaration(LLVMCodeGenerator *self, Declaration *decl);

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt);

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt);

LLVMValueRef genBinaryExpression(LLVMCodeGenerator *self, BinaryExpr *expr);

LLVMValueRef genExpression(LLVMCodeGenerator *self, Expression *expr);

// Definitions

LLVMCodeGenerator *createLLVMCodeGenerator(Vector *sourceFiles) {
	LLVMCodeGenerator *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->builder = LLVMCreateBuilder();
	return self;
}

void destroyLLVMCodeGenerator(LLVMCodeGenerator *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		LLVMDisposeModule(sourceFile->module);
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files iteration %d", i);
	}

	free(self);
	verboseModeMessage("Destroyed compiler");
}

static void consumeAstNode(LLVMCodeGenerator *self) {
	self->currentNode += 1;
}

static void consumeAstNodeBy(LLVMCodeGenerator *self, int amount) {
	self->currentNode += amount;
}

LLVMValueRef genBinaryExpression(LLVMCodeGenerator *self, BinaryExpr *expr) {
	LLVMValueRef lhs = genExpression(self, expr->lhand);
	LLVMValueRef rhs = genExpression(self, expr->lhand);
	if (!lhs || !rhs) {
		genError("Invalid expression");
		// TODO
	}

	if (!strcmp(expr->binaryOp, "+")) {
		return LLVMBuildFAdd(self->builder, lhs, rhs, "add");
	}
	else if (!strcmp(expr->binaryOp, "-")) {
		return LLVMBuildFSub(self->builder, lhs, rhs, "sub");
	}
	else if (!strcmp(expr->binaryOp, "*")) {
		return LLVMBuildFMul(self->builder, lhs, rhs, "mul");
	}
	else if (!strcmp(expr->binaryOp, "/")) {
		return LLVMBuildFDiv(self->builder, lhs, rhs, "div");
	}
}

LLVMValueRef genFunctionCall(LLVMCodeGenerator *self, Call *call) {
	char *funcName = getVectorItem(call->callee, 0);
	LLVMValueRef func = LLVMGetNamedFunction(self->currentSourceFile->module, funcName);

	if (!func) {
		genError("Function not found in module");
		return false;
	}

	if (LLVMCountParams(func) != call->arguments->size) {
		genError("Function has too many/few arguments");
		return false;
	}

	LLVMValueRef *args = malloc(sizeof(LLVMValueRef) * call->arguments->size);
	for (int i = 0; i < call->arguments->size; i++) {
		args[i] = genExpression(self, getVectorItem(call->arguments, i));
		if (!args[i]) {
			genError("Could not evaluate argument in function call %s", funcName);
			free(args);
			return false;
		}
	}

	LLVMBuildCall(self->builder, func, args, call->arguments->size, "calltmp");
}

LLVMValueRef genTypeName(LLVMCodeGenerator *self, TypeName *name) {
	printf("todo\n");
	return false;
}

LLVMValueRef genLiteral(LLVMCodeGenerator *self, Literal *lit) {
	switch (lit->type) {
		case INT_LITERAL_NODE: 
			printf("%d\n", lit->intLit->value);
			return LLVMConstInt(LLVMInt32Type(), lit->intLit->value, false);
		default:
			printf("hmm?\n");
			break;
	}
	return false;
}

LLVMValueRef genTypeLit(LLVMCodeGenerator *self, TypeLit *lit) {
	switch (lit->type) {
		default:
			printf("%s\n", getNodeTypeName(lit->type));
			break;
	}
}

LLVMValueRef genType(LLVMCodeGenerator *self, Type *type) {
	switch (type->type) {
		case TYPE_NAME_NODE:
			genTypeName(self, type->typeName);
			break;
		case TYPE_LIT_NODE:
			genTypeLit(self, type->typeLit);
			break;
	}
}

LLVMValueRef genExpression(LLVMCodeGenerator *self, Expression *expr) {
	switch (expr->exprType) {
		case TYPE_NODE: return genType(self, expr->type);
		case LITERAL_NODE: return genLiteral(self, expr->lit);
		case BINARY_EXPR_NODE: return genBinaryExpression(self, expr->binary);
		case UNARY_EXPR_NODE: printf("unary\n"); break;
		case FUNCTION_CALL_NODE: return genFunctionCall(self, expr->call);
		case ARRAY_INITIALIZER_NODE: printf("array init\n"); break;
		case ARRAY_INDEX_NODE: printf("array index\n"); break;
		case ALLOC_NODE: printf("alloc\n"); break;
		case SIZEOF_NODE: printf("sizeof\n"); break;
		default:
			errorMessage("Unknown node in expression %d", expr->exprType);
			break;
	}
}

LLVMValueRef genFunctionSignature(LLVMCodeGenerator *self, FunctionSignature *decl) {
	// store arguments from func signature
	unsigned int argCount = decl->parameters->paramList->size;

	// lookup func
	LLVMValueRef func = LLVMGetNamedFunction(self->currentSourceFile->module, decl->name);
	if (func) {
		if (LLVMCountParams(func) != argCount) {
			genError("Function exists with different function signature");
			return false;
		}
	}
	else {
		// set llvm params
		LLVMTypeRef *params = safeMalloc(sizeof(LLVMTypeRef) * argCount);
		for (int i = 0; i < argCount; i++) {
			ParameterSection *param = getVectorItem(decl->parameters->paramList, i);
			if (param->type->type != TYPE_NAME_NODE) {
				genError("Unsupported type :(");
				return false;
			}
			int type = getTypeFromString(param->type->typeName->name);
			params[i] = getLLVMType(type);
		}

		// create func prototype and add it to the module
		int funcType = getTypeFromString(decl->type->typeName->name);
		LLVMTypeRef funcTypeRef = LLVMFunctionType(getLLVMType(funcType), params, argCount, false);
		func = LLVMAddFunction(self->currentSourceFile->module, decl->name, funcTypeRef);
	}

	return func;
}

LLVMValueRef genStatement(LLVMCodeGenerator *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			return genUnstructuredStatementNode(self, stmt->unstructured);
		case STRUCTURED_STATEMENT_NODE:
			return genStructuredStatementNode(self, stmt->structured);
		case MACRO_NODE: 
			printf("macro ignored\n");
			consumeAstNode(self); 
			break; // ignore the macro
		default:
			errorMessage("Unknown statement %s\n", getNodeTypeName(stmt->type));
			break;
	}
	printf("returned null idk why\n");
	return false;
}

LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl) {
	LLVMValueRef prototype = genFunctionSignature(self, decl->signature);
	if (!prototype) {
		genError("hmm");
		return NULL;
	}

	LLVMBasicBlockRef block = LLVMAppendBasicBlock(prototype, "entry");
	LLVMPositionBuilderAtEnd(self->builder, block);

	for (int i = 0; i < decl->body->stmtList->stmts->size; i++) {
		LLVMValueRef body = genStatement(self, getVectorItem(decl->body->stmtList->stmts, i));
	}

	return prototype;
}

LLVMValueRef genDeclaration(LLVMCodeGenerator *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: return genFunctionDecl(self, decl->funcDecl);
	}
}

LLVMValueRef genLeaveStatNode(LLVMCodeGenerator *self, LeaveStat *leave) {
	switch (leave->type) {
		case RETURN_STAT_NODE: {
			LLVMValueRef expr = NULL;
			if (leave->retStmt->expr) {
				expr = genExpression(self, leave->retStmt->expr);
			}
			LLVMBuildRet(self->builder, expr != NULL ? expr : LLVMVoidType());
			break;
		}
	}
}

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
		case DECLARATION_NODE: return genDeclaration(self, stmt->decl);
		case EXPR_STAT_NODE: return genExpression(self, stmt->expr); 
		case LEAVE_STAT_NODE: return genLeaveStatNode(self, stmt->leave);
		case FUNCTION_CALL_NODE: return genFunctionCall(self, stmt->call);
		default: 
			printf("found %s\n", getNodeTypeName(stmt->type));
			break;
	}
	return false;
}

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt) {
	switch (stmt->type) {
		default:
			printf("omg!\n");
			break;
	}
	return false;
}

void traverseAST(LLVMCodeGenerator *self) {
	for (int i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *stmt = getVectorItem(self->abstractSyntaxTree, i);
		genStatement(self, stmt);
	}
}

void startLLVMCodeGeneration(LLVMCodeGenerator *self) {
	Vector *asmFiles = createVector(VECTOR_EXPONENTIAL);
	
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		sf->module = LLVMModuleCreateWithName(sf->name);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		traverseAST(self);

		// just dump mods for now
		LLVMDumpModule(sf->module);
		
		sds bitcodeFilename = sdsempty();
		bitcodeFilename = sdscat(bitcodeFilename, sf->name);
		bitcodeFilename = sdscat(bitcodeFilename, ".bc");
		
		char *error = NULL;
		int verify_result = LLVMVerifyModule(sf->module, LLVMReturnStatusAction, &error);
		if (verify_result) {
			genError("%s", error);
		} else {
			if (LLVMWriteBitcodeToFile(sf->module, bitcodeFilename)) {
				genError("Failed to write bit-code");
			}
		}
		LLVMDisposeMessage(error);
		
		sds asmFilename = sdsempty();
		asmFilename = sdscat(asmFilename, sf->name);
		asmFilename = sdscat(asmFilename, ".s");
		
		// convert bitcode files to assembly files
		sds toasmcommand = sdsempty();
		toasmcommand = sdscat(toasmcommand, "llc ");
		toasmcommand = sdscat(toasmcommand, bitcodeFilename);
		toasmcommand = sdscat(toasmcommand, " -o ");
		toasmcommand = sdscat(toasmcommand, asmFilename);
		
		int toasmresult = system(toasmcommand);
		if (toasmresult)
			genError("Couldn't assemble bitcode file %s", bitcodeFilename);
		
		int removeBitcodeResult = remove(bitcodeFilename);
		if (removeBitcodeResult)
			genError("Couldn't remove bitcode file %s", bitcodeFilename);
		
		pushBackItem(asmFiles, asmFilename);
		
		sdsfree(bitcodeFilename);
		sdsfree(toasmcommand);
	}
	
	sds linkCommand = sdsempty();
	linkCommand = sdscat(linkCommand, COMPILER);
	linkCommand = sdscat(linkCommand, " ");
	
	// get all the asm files to compile
	for (int i = 0; i < asmFiles->size; i++) {
		linkCommand = sdscat(linkCommand, getVectorItem(asmFiles, i));
		linkCommand = sdscat(linkCommand, " ");
	}
	
	linkCommand = sdscat(linkCommand, " -o ");
	linkCommand = sdscat(linkCommand, OUTPUT_EXECUTABLE_NAME);
	
	int linkResult = system(linkCommand);
	if (linkResult)
		genError("Couldn't link object files");
	
	sdsfree(linkCommand);
	
	for (int i = 0; i < asmFiles->size; i++) {
		int freeAsmResult = remove(getVectorItem(asmFiles, i));
		if (freeAsmResult)
			genError("Couldn't remove assembly file %s", getVectorItem(asmFiles, i));
		sdsfree(getVectorItem(asmFiles, i));
	}
	
	destroyVector(asmFiles);
}

static LLVMTypeRef getIntType() {
	switch (sizeof(int)) {
		case 2: return LLVMInt16Type();
		case 4: return LLVMInt32Type();
		case 8: return LLVMInt64Type();
		default:
			// either something fucked up, or we're in the future on 128 bit machines
			verboseModeMessage("You have some wacky-sized int type, switching to 16 bit for default!");
			return LLVMInt16Type();
	}
}

static LLVMTypeRef getLLVMType(DataType type) {
	switch (type) {
		// case INT_128_TYPE:
		// case UINT_128_TYPE:
		// 	return LLVMIntType(128);
			
		case INT_64_TYPE:
		case UINT_64_TYPE:
			return LLVMInt64Type();
			
		case INT_32_TYPE:
		case UINT_32_TYPE:
			return LLVMInt32Type();
			
		case INT_16_TYPE:
		case UINT_16_TYPE:
			return LLVMInt16Type();
			
		case INT_8_TYPE:
		case UINT_8_TYPE:
			return LLVMInt8Type();
		
		case FLOAT_128_TYPE:
			return LLVMFP128Type();
		
		case FLOAT_64_TYPE:
			return LLVMDoubleType();
			
		case FLOAT_32_TYPE:
			return LLVMFloatType();
			
		case INT_TYPE:
			return getIntType();
			
		case BOOL_TYPE:
			return LLVMInt1Type();
			
		case CHAR_TYPE:
			genError("Char type unimplemented");
			return NULL; // gonna get replaced
			
		case VOID_TYPE:
			return LLVMVoidType();
			
		case UNKNOWN_TYPE:
			genError("Unknown type");
			return NULL;
	}
}
