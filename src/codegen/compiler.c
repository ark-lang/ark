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

		int i;
		for (i = 0; i < self->abstractSyntaxTree->size; i++) {
			statementDriver(self, getVectorItem(self->abstractSyntaxTree, i));
		}
	}
}

void declarationDriver(Compiler *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: generateFunctionCode(self, decl->funcDecl); break;
	}
}

void unstructuredStatementDriver(Compiler *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
		case DECLARATION_NODE: declarationDriver(self, stmt->decl); break;
	}
}

void structuredStatementDriver(Compiler *self, StructuredStatement *stmt) {
	switch (stmt->type) {

	}
}

void statementDriver(Compiler *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			unstructuredStatementDriver(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE: {
			structuredStatementDriver(self, stmt->structured);
			break;
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
