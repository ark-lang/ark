#include "semantic.h"

#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles) {
	SemanticAnalyzer *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->sourceFiles = sourceFiles;
	self->abstractSyntaxTree = NULL;
	self->currentSourceFile = NULL;
	self->currentNode = 0;
	self->failed = false;
	return self;
}

void analyzeBlock(SemanticAnalyzer *self, Block *block) {
	for (int i = 0; i < block->stmtList->stmts->size; i++) {
		Statement *stmt = getVectorItem(block->stmtList->stmts, i);
		analyzeStatement(self, stmt);
	}
}

void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl) {
	FunctionDecl *mapDecl = NULL;	
	if (hashmap_get(self->funcSymTable, decl->signature->name, (void**) &mapDecl) == MAP_MISSING) {
		hashmap_put(self->funcSymTable, decl->signature->name, decl);
		if (!decl->prototype) {
			analyzeBlock(self, decl->body);
		}
	}
	else {
		semanticError("Function with name `%s` already exists", decl->signature->name);
	}
}

void analyzeDeclaration(SemanticAnalyzer *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: analyzeFunctionDeclaration(self, decl->funcDecl); break;
	}
}

void analyzeFunctionCall(SemanticAnalyzer *self, Call *call) {
	char *callee = getVectorItem(call->callee, 0);

	FunctionDecl *decl = NULL;	
	// check the function we're calling exists
	if (hashmap_get(self->funcSymTable, callee, (void**) &decl) == MAP_MISSING) {
		semanticError("Attempting to call undefined function `%s`", callee);
	}
	// it exists, check arguments match in length
	else {
		int argsNeeded = decl->signature->parameters->paramList->size;
		int argsGot = call->arguments->size;
		char *callee = getVectorItem(call->callee, 0); // FIXME

		if (argsGot > argsNeeded) {
			semanticError("Too many arguments to function `%s`", callee);
		}
		else if (argsGot < argsNeeded) {
			semanticError("Too few arguments to function `%s`", callee);
		}
	}
}

void analyzeUnstructuredStatement(SemanticAnalyzer *self, UnstructuredStatement *unstructured) {
	switch (unstructured->type) {
		case DECLARATION_NODE: analyzeDeclaration(self, unstructured->decl); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, unstructured->call); break;
	}
}

void analyzeStructuredStatement(SemanticAnalyzer *self, StructuredStatement *structured) {
	ALLOY_UNUSED_OBJ(self);
	ALLOY_UNUSED_OBJ(structured);	
}

void analyzeStatement(SemanticAnalyzer *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			analyzeUnstructuredStatement(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE: 
			analyzeStructuredStatement(self, stmt->structured);
			break;
		default:
			semanticError("WHAT YEAR IS IT?");
			break;
	}
}

void checkMainExists(SemanticAnalyzer *self) {
	FunctionDecl *main = NULL;
	if (hashmap_get(self->funcSymTable, MAIN_FUNC, (void**) &main) == MAP_MISSING) {
		semanticError("Undefined reference to `main`");
		self->failed = true;
	}
}

void startSemanticAnalysis(SemanticAnalyzer *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;
		
		for (int j = 0; j < self->abstractSyntaxTree->size; j++) {
			Statement *stmt = getVectorItem(self->abstractSyntaxTree, j);
			analyzeStatement(self, stmt);
		}
	}

	checkMainExists(self);
}

void destroySemanticAnalyzer(SemanticAnalyzer *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	free(self);
}
