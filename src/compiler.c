#include "compiler.h"

Compiler *createCompiler(Vector *sourceFiles) {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->functions = hashmap_new();
	self->structures = hashmap_new();
	return self;
}

void emitCode(Compiler *self, char *fmt, ...) {
	va_list args;
	va_start(args, fmt);
	vfprintf(self->currentSourceFile->outputFile, fmt, args);
	va_end(args);
}

void emitExpression(Compiler *self, Expression *expr) {

}

void emitType(Compiler *self, Type *type) {

}

void emitParameters(Compiler *self, Parameters *params) {

}

void emitReceiver(Compiler *self, Receiver *rec) {

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
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sourceFile;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		writeSourceFile(self->currentSourceFile);

		// compile code
		compileAST(self);

		// close files
		closeSourceFile(self->currentSourceFile);
	}
	sds buildCommand = sdsempty();

	// append the compiler to use etc
	buildCommand = sdscat(buildCommand, COMPILER);
	buildCommand = sdscat(buildCommand, " ");
	buildCommand = sdscat(buildCommand, ADDITIONAL_COMPILER_ARGS);
	buildCommand = sdscat(buildCommand, " ");

	// append the filename to the build string
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		buildCommand = sdscat(buildCommand, sourceFile->generatedSourceName);
	}

	buildCommand = sdscat(buildCommand, " -o ");
	buildCommand = sdscat(buildCommand, OUTPUT_EXECUTABLE_NAME);

	// just for debug purposes
	verboseModeMessage("running cl args: `%s`", buildCommand);
	system(buildCommand);
	sdsfree(buildCommand); // deallocate dat shit baby
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
	// now we can destroy stuff
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files on %d iteration.", i);
	}

	hashmap_free(self->functions);
	hashmap_free(self->structures);
	free(self);
	verboseModeMessage("Destroyed compiler");
}
