#include "sourcefile.h"

SourceFile *createSourceFile(char *fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->headerFile = createHeaderFile(fileName);
	sourceFile->name = getFileName(sourceFile->fileName);
	sourceFile->alloyFileContents = readFile(fileName);

	return sourceFile;
}

void writeFiles(SourceFile *sourceFile) {
	writeSourceFile(sourceFile);
	writeHeaderFile(sourceFile->headerFile);
}

void writeSourceFile(SourceFile *sourceFile) {
	sourceFile->outputFile = fopen(sourceFile->name ".c", "w");
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
	free(sourceFile->name);
}

void destroySourceFile(SourceFile *sourceFile) {
	if (sourceFile) {
		destroyHeaderFile(sourceFile->headerFile);
		free(sourceFile->alloyFileContents);
		free(sourceFile);
	}
}
