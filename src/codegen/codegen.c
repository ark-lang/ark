#include "codegen.h"

char *BOILERPLATE =
"#include <stdlib.h>\n"
"#include <stdio.h>\n"
"#include <stdbool.h>\n"
"\n"
"typedef char *str;" CC_NEWLINE
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

CodeGenerator *createCodeGenerator(Vector *sourceFiles) {
	CodeGenerator *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->symtable = hashmap_new();
	return self;
}

void emitCode(CodeGenerator *self, char *fmt, ...) {
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

void consumeAstNode(CodeGenerator *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(CodeGenerator *self, int amount) {
	self->currentNode += amount;
}

void traverseAST(CodeGenerator *self) {

}

void startCodeGeneration(CodeGenerator *self) {
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;
		
		writeFiles(self->currentSourceFile);

		self->writeState = WRITE_SOURCE_STATE;
		// _gen_name.h is the typical name for the headers and c files that are generated
		emitCode(self, "#include \"_gen_%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "#ifndef __%s_H\n", self->currentSourceFile->name);
		emitCode(self, "#define __%s_H\n\n", self->currentSourceFile->name);

		emitCode(self, BOILERPLATE);

		// compile code
		traverseAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", self->currentSourceFile->name);

		// close files
		closeFiles(self->currentSourceFile);
	}

	sds buildCommand = sdsempty();

	// append the compiler to use etc
	buildCommand = sdscat(buildCommand, COMPILER);
	buildCommand = sdscat(buildCommand, " ");
	buildCommand = sdscat(buildCommand, ADDITIONAL_COMPILER_ARGS);
	buildCommand = sdscat(buildCommand, " -o ");
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
	verboseModeMessage("running cl args: `%s`", buildCommand);
	system(buildCommand);
	sdsfree(buildCommand); // deallocate dat shit baby
}

void destroyCodeGenerator(CodeGenerator *self) {
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
