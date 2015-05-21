#include "sourcefile.h"

SourceFile *createSourceFile(sds fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
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

void destroySourceFile(SourceFile *sourceFile) {
	verboseModeMessage("Destroyed source file `%s`", sourceFile->name);

	if (sourceFile->tokens) destroyVector(sourceFile->tokens); // TODO destroy elements in tokens
	if (sourceFile->ast) destroyParseTree(sourceFile->ast);

	sdsfree(sourceFile->fileName);
	free(sourceFile->name);
	free(sourceFile->fileContents); 
	free(sourceFile);
}
