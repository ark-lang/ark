#include "sourcefile.h"

SourceFile *createSourceFile(char *fileName, Vector *nodes) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->fileContents = malloc(sizeof(char));
	sourceFile->fileContents[0] = '\0';
	sourceFile->nodes = nodes;

	strcpy(sourceFile->outputFileName, fileName);
	str_append(sourceFile->outputFileName, ".c");
	printf("%s\n", sourceFile->outputFileName);

	sourceFile->headerFile = createHeaderFile(fileName);

	return sourceFile;
}

void writeFiles(SourceFile *sourceFile) {
	writeSourceFile(sourceFile);
	writeHeaderFile(sourceFile->headerFile);
}

void writeSourceFile(SourceFile *sourceFile) {
	FILE *file = fopen(sourceFile->outputFileName, "w");
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
		remove(sourceFile->outputFileName);
		destroyHeaderFile(sourceFile->headerFile);
		free(sourceFile->fileContents);
		free(sourceFile);
	}
}