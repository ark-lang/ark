#include "compiler.h"

Compiler *createCompiler(Vector *sourceFiles) {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->symtable = hashmap_new();
	return self;
}

void emitExpression(Compiler *self, Expression *expr) {

}

void emitType(Compiler *self, Type *type) {

}

void emitParameters(Compiler *self, Parameters *params) {

}

void emitReceiver(Compiler *self, Receiver *rec) {

}

void emitFunctionPrologue(Compiler *self) {

}

void emitFunctionSignature(Compiler *self, FunctionSignature *func) {

}

void emitStructuredStatement(Compiler *self, StructuredStatement *stmt) {

}

void emitUnstructuredStatement(Compiler *self, UnstructuredStatement *stmt) {

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

}

void emitIdentifierList(Compiler *self, IdentifierList *list) {

}

void emitFieldDeclList(Compiler *self, FieldDeclList *list) {

}

void emitStructDecl(Compiler *self, StructDecl *decl) {

}

void emitDeclaration(Compiler *self, Declaration *decl) {

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
