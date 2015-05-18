#include "sourcefile.h"

SourceFile *createSourceFile(sds fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->headerFile = createHeaderFile(fileName);
	sourceFile->name = getFileName(sourceFile->fileName);
	sourceFile->fileContents = readFile(fileName);
	sourceFile->tokens = NULL;
	sourceFile->ast = NULL;

	if (!sourceFile->fileContents) {
		errorMessage("Failed to read file %s", sourceFile->fileName);
		destroySourceFile(sourceFile);
		return NULL;
	}

	return sourceFile;
}

void writeFiles(SourceFile *sourceFile) {
	writeSourceFile(sourceFile);
	writeHeaderFile(sourceFile->headerFile);
}

void writeSourceFile(SourceFile *sourceFile) {
	sourceFile->generatedSourceName = sdsempty();
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, "_gen_");
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, sourceFile->name);
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, ".c");

	verboseModeMessage("Generated source file `%s` as `%s`", sourceFile->name, sourceFile->generatedSourceName);
	sourceFile->outputFile = fopen(sourceFile->generatedSourceName, "w");
	if (!sourceFile->outputFile) {
		perror("fopen: failed to open file");
		return;
	}
}

void closeFiles(SourceFile *sourceFile) {
	closeSourceFile(sourceFile);
	closeHeaderFile(sourceFile->headerFile);
}

void closeSourceFile(SourceFile *sourceFile) {
	fclose(sourceFile->outputFile);
}

void destroySourceFile(SourceFile *sourceFile) {
	#ifdef ENABLE_LLVM
	if (!OUTPUT_C && !LLVM_CODEGEN) remove(sourceFile->generatedSourceName);
	#else
	if (!OUTPUT_C) remove(sourceFile->generatedSourceName);
	#endif

	destroyHeaderFile(sourceFile->headerFile);

	verboseModeMessage("Destroyed source file `%s`", sourceFile->name);
	if (sourceFile->tokens) destroyVector(sourceFile->tokens); // TODO destroy elements in tokens
	if (sourceFile->ast) destroyParseTree(sourceFile->ast);
	sdsfree(sourceFile->fileName);
	free(sourceFile->name);
	#ifndef ENABLE_LLVM
	sdsfree(sourceFile->generatedSourceName);
	#endif
	free(sourceFile->fileContents); 
	free(sourceFile);
}
