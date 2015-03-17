#include "sourcefile.h"

SourceFile *createSourceFile(char *fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->fileContents = readFile(fileName);
	sourceFile->headerFile = createHeaderFile(fileName);
	return sourceFile;
}

void writeFiles(SourceFile *sourceFile) {
	writeSourceFile(sourceFile);
	writeHeaderFile(sourceFile->headerFile);
}

void writeSourceFile(SourceFile *sourceFile) {
	FILE *file = fopen(JOIN_STR(sourceFile->fileName, ".c"), "w");
	if (!file) {
		perror("fopen: failed to open file");
		return;
	}

	fprintf(file, "%s", sourceFile->fileContents);
	fclose(file);

	system(JOIN_STR(COMPILER, sourceFile->outputFileName));
}

void destroySourceFile(SourceFile *sourceFile) {
	if (sourceFile) {
		remove(JOIN_STR(sourceFile->fileName, ".c"));
		destroyHeaderFile(sourceFile->headerFile);
		free(sourceFile->fileContents);
		free(sourceFile);
	}
}
