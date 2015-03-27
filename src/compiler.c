#include "compiler.h"

char *BOILERPLATE =
"#include <stdlib.h>\n"
"#include <stdbool.h>\n"
"\n"
"typedef char *string;" CC_NEWLINE
"typedef unsigned long long u64;" CC_NEWLINE
"typedef unsigned int u32;" CC_NEWLINE
"typedef unsigned short u16;" CC_NEWLINE
"typedef unsigned char u8;" CC_NEWLINE
"typedef long long s64;" CC_NEWLINE
"typedef int s32;" CC_NEWLINE
"typedef short s16;" CC_NEWLINE
"typedef char s8;" CC_NEWLINE
"typedef float f32;" CC_NEWLINE
"typedef double f64;" CC_NEWLINE
;

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

	switch (self->writeState) {
		case WRITE_SOURCE_STATE:
			vfprintf(self->currentSourceFile->outputFile, fmt, args);
			va_end(args);
			break;
		case WRITE_HEADER_STATE:
			vfprintf(self->currentSourceFile->headerFile->outputFile, fmt, args);
			va_end(args);
			break;
	}
}

void emitType(Compiler *self, Type *type) {
	switch (type->type) {
	case POINTER_TYPE_NODE: break;
	case ARRAY_TYPE_NODE: break;
	case TYPE_NAME_NODE:
		emitCode(self, "%s", type->typeName->name);
		break;
	}
}

void emitFunctionDecl(Compiler *self, FunctionDecl *decl) {
	self->writeState = WRITE_HEADER_STATE;
	emitType(self, decl->signature->type);
	printf("emit a function decl, idk?");
}

void emitDeclaration(Compiler *self, Declaration *decl) {
	switch (decl->declType) {
	case FUNC_DECL: emitFunctionDecl(self, decl->funcDecl); break;
	}
}

void emitUnstructuredStat(Compiler *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
	case DECLARATION_NODE: emitDeclaration(self, stmt->decl); break;
	}
}

void emitStructuredStat(Compiler *self, StructuredStatement *stmt) {

}

void emitStatement(Compiler *self, Statement *stmt) {

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

		writeFiles(self->currentSourceFile);

		self->writeState = WRITE_SOURCE_STATE;
		// _gen_name.h is the typical name for the headers and c files that are generated
		emitCode(self, "#include \"_gen_%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		sds nameInUpperCase = toUppercase(self->currentSourceFile->name);
		if (!nameInUpperCase) {
			errorMessage("Failed to convert case to upper");
			return;
		}

		emitCode(self, "#ifndef __%s_H\n", nameInUpperCase);
		emitCode(self, "#define __%s_H\n\n", nameInUpperCase);

		emitCode(self, BOILERPLATE);

		// compile code
		compileAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", nameInUpperCase);

		sdsfree(nameInUpperCase);

		// close files
		closeFiles(self->currentSourceFile);
	}

	sds buildCommand = sdsempty();

	// append the compiler to use etc
	buildCommand = sdscat(buildCommand, COMPILER);
	buildCommand = sdscat(buildCommand, " -std=c99 -Wall -o ");
	buildCommand = sdscat(buildCommand, OUTPUT_EXECUTABLE_NAME);
	buildCommand = sdscat(buildCommand, " ");

	// append the filename to the build string
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		buildCommand = sdscat(buildCommand, sourceFile->generatedSourceName);

		if (i != self->sourceFiles->size - 1) // stop whitespace at the end!
			buildCommand = sdscat(buildCommand, " ");
	}

	// just for debug purposes
	debugMessage("running cl args: `%s`", buildCommand);
	system(buildCommand);
	sdsfree(buildCommand); // deallocate dat shit baby
}

void compileAST(Compiler *self) {
	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *currentStmt = getVectorItem(self->abstractSyntaxTree, i);

		switch (currentStmt->type) {
		case UNSTRUCTURED_STMT: emitUnstructuredStat(self, currentStmt->unstructured); break;
		case STRUCTURED_STMT: emitStructuredStat(self, currentStmt->structured); break;
		default:
			printf("idk?\n");
			break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	// now we can destroy stuff
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		// don't call destroyHeaderFile since it's called in this function!!!!
		destroySourceFile(sourceFile);
		debugMessage("Destroyed source files on %d iteration.", i);
	}

	hashmap_free(self->functions);
	hashmap_free(self->structures);
	free(self);
	debugMessage("Destroyed compiler");
}
