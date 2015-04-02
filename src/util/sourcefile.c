#include "sourcefile.h"

SourceFile *createSourceFile(sds fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->name = getFileName(sourceFile->fileName);
	sourceFile->alloyFileContents = readFile(fileName);

	if (!sourceFile->alloyFileContents) {
		errorMessage("Failed to read file %s", sourceFile->fileName);
		destroySourceFile(sourceFile);
		return NULL;
	}

	return sourceFile;
}

void destroySourceFile(SourceFile *sourceFile) {
	verboseModeMessage("Destroyed Source File `%s`", sourceFile->name);
	sdsfree(sourceFile->fileName);
	free(sourceFile->name);
	free(sourceFile->alloyFileContents); 
	free(sourceFile);
}
