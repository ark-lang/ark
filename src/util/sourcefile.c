#include "sourcefile.h"

SourceFile *createSourceFile(sds fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->headerFile = createHeaderFile(fileName);
	sourceFile->name = getFileName(sourceFile->fileName);
	sourceFile->alloyFileContents = readFile(fileName);
	sourceFile->tokens = NULL;
	sourceFile->ast = NULL;

	if (!sourceFile->alloyFileContents) {
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
	if (!OUTPUT_C) remove(sourceFile->generatedSourceName);

	destroyHeaderFile(sourceFile->headerFile);

	verboseModeMessage("Destroyed Source File `%s`", sourceFile->name);
	if (sourceFile->tokens) destroyVector(sourceFile->tokens);
	if (sourceFile->ast) destroyVector(sourceFile->ast);
	sdsfree(sourceFile->fileName);
	free(sourceFile->name);
	sdsfree(sourceFile->generatedSourceName);
	free(sourceFile->alloyFileContents); 
	free(sourceFile);
}
