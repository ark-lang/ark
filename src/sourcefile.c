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

void writeSourceFile(SourceFile *sourceFile) {
	sourceFile->generatedSourceName = sdsempty();
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, OUTPUT_PREFIX);
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, sourceFile->name);
	sourceFile->generatedSourceName = sdscat(sourceFile->generatedSourceName, OUTPUT_EXTENSION);

	verboseModeMessage("Generated source file `%s` as `%s`", sourceFile->name, sourceFile->generatedSourceName);
	sourceFile->outputFile = fopen(sourceFile->generatedSourceName, "w");
	if (!sourceFile->outputFile) {
		perror("fopen: failed to open file");
		return;
	}
}

void closeSourceFile(SourceFile *sourceFile) {
	fclose(sourceFile->outputFile);
}

void destroySourceFile(SourceFile *sourceFile) {
	if (!OUTPUT_C) remove(sourceFile->generatedSourceName);

	verboseModeMessage("Destroyed Source File `%s`", sourceFile->name);
	sdsfree(sourceFile->fileName);
	free(sourceFile->name);
	sdsfree(sourceFile->generatedSourceName);
	free(sourceFile->alloyFileContents); 
	free(sourceFile);
}
