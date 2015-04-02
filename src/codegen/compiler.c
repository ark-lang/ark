#include "compiler.h"

Compiler *createCompiler(Vector *sourceFiles) {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->symtable = hashmap_new();

	// initialize llvm shit
	self->module = LLVMModuleCreateWithName("jacks dog");
	self->builder = LLVMCreateBuilder();
	return self;
}

LLVMTypeRef getTypeRef(Type *type) {
	if (type->type == TYPE_NAME_NODE) {
		int dataType = getTypeFromString(type->typeName->name);
		switch (dataType) {
		case INT_64_TYPE: return LLVMInt64Type();
		case INT_32_TYPE: return LLVMInt32Type();
		case INT_16_TYPE: return LLVMInt16Type();
		case INT_8_TYPE: return LLVMInt8Type();
		case VOID_TYPE: return LLVMVoidType();
		}
	}

	return false;
}

void emitExpression(Compiler *self, Expression *expr) {

}

void emitType(Compiler *self, Type *type) {

}

void emitParameters(Compiler *self, Parameters *params) {

}

void emitReceiver(Compiler *self, Receiver *rec) {

}

void emitStructuredStatement(Compiler *self, StructuredStatement *stmt) {

}

void emitUnstructuredStatement(Compiler *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
	case DECLARATION_NODE:
		emitDeclaration(self, stmt->decl);
		break;
	}
}

void emitBlock(Compiler *self, Block *block) {

}

void emitForStat(Compiler *self, ForStat *stmt) {

}

void emitIfStat(Compiler *self, IfStat *stmt) {

}

void emitMatchStat(Compiler *self, MatchStat *stmt) {

}

void emitStatementList(Compiler *self, StatementList *stmtList) {

}

void emitFunctionDecl(Compiler *self, FunctionDecl *decl) {
	const int numOfParams = decl->signature->parameters->paramList->size;

	LLVMTypeRef params[numOfParams];

	int i;
	for (i = 0; i < numOfParams; i++) {
		ParameterSection *param = getVectorItem(decl->signature->parameters->paramList, i);
		LLVMTypeRef type = getTypeRef(param->type);
		if (type) {
			params[i] = type;
		}
	}

	LLVMTypeRef returnType = getTypeRef(decl->signature->type);
	if (returnType) {
		LLVMTypeRef funcRet = LLVMFunctionType(returnType, params, numOfParams, false);
		LLVMValueRef func = LLVMAddFunction(self->module, decl->signature->name, funcRet);
	}
}

void emitIdentifierList(Compiler *self, IdentifierList *list) {

}

void emitFieldDeclList(Compiler *self, FieldDeclList *list) {

}

void emitStructDecl(Compiler *self, StructDecl *decl) {

}

void emitDeclaration(Compiler *self, Declaration *decl) {
	switch (decl->type) {
	case FUNCTION_DECL_NODE:
		emitFunctionDecl(self, decl->funcDecl);
		break;
	case STRUCT_DECL_NODE:
		emitStructDecl(self, decl->structDecl);
		break;
	case VARIABLE_DECL_NODE:
		break;
	}
}

void consumeAstNode(Compiler *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(Compiler *self, int amount) {
	self->currentNode += amount;
}

void startCompiler(Compiler *self) {
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		compileAST(self);
	}
}

void compileAST(Compiler *self) {
	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *currentStmt = getVectorItem(self->abstractSyntaxTree, i);

		switch (currentStmt->type) {
			case UNSTRUCTURED_STATEMENT_NODE: emitUnstructuredStatement(self, currentStmt->unstructured); break;
			case STRUCTURED_STATEMENT_NODE: emitStructuredStatement(self, currentStmt->structured); break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files on %d iteration.", i);
	}

	hashmap_free(self->symtable);
	free(self);
	verboseModeMessage("Destroyed compiler");
}
