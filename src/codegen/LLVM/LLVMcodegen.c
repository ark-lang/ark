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

void genDeclaration(LLVMCodeGenerator *self, Declaration *decl);

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt);

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt);

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
	}
	return false;
}

LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl) {
	LLVMValueRef prototype = genFunctionSignature(self, decl->signature);
	if (!prototype) {
		return NULL;
	}

	LLVMBasicBlockRef block = LLVMAppendBasicBlock(prototype, "entry");
	LLVMPositionBuilderAtEnd(self->builder, block);

	LLVMValueRef body = getIntType(); // genStatement(self, decl->body);
	if (!body) {
		LLVMDeleteFunction(prototype);
		return false;
	}

	LLVMBuildRet(self->builder, body);
	if (LLVMVerifyFunction(prototype, LLVMPrintMessageAction)) {
		genError("Invalid function");
		LLVMDeleteFunction(prototype);
		return false;
	}
}

void genDeclaration(LLVMCodeGenerator *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: genFunctionDecl(self, decl->funcDecl); break;
	}
}

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
		case DECLARATION_NODE: genDeclaration(self, stmt->decl); break;
		case EXPR_STAT_NODE: printf("idk\n"); break;
	}
}

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt) {}

void traverseAST(LLVMCodeGenerator *self) {
	for (int i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *stmt = getVectorItem(self->abstractSyntaxTree, i);
		genStatement(self, stmt);
	}
}

void startLLVMCodeGeneration(LLVMCodeGenerator *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		sf->module = LLVMModuleCreateWithName(sf->name);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		traverseAST(self);

		LLVMDumpModule(sf->module);
	}
}

static LLVMTypeRef getIntType() {
	switch (sizeof(int)) {
		case 2: printf("16\n"); return LLVMInt16Type();
		case 4: printf("32\n"); return LLVMInt32Type();
		case 8: printf("64\n"); return LLVMInt64Type();
		default:
			// either something fucked up, or we're in the future on 128 bit machines
			verboseModeMessage("You have some wacky-sized int type, switching to 16 bit for default!");
			return LLVMInt16Type();
	}
}

static LLVMTypeRef getLLVMType(DataType type) {
	switch (type) {
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
