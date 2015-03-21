#include "sourcefile.h"

SourceFile *createSourceFile(sds fileName) {
	SourceFile *sourceFile = malloc(sizeof(*sourceFile));
	sourceFile->fileName = fileName;
	sourceFile->headerFile = createHeaderFile(fileName);
	sourceFile->name = getFileName(sourceFile->fileName);
	sourceFile->alloyFileContents = readFile(fileName);
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
	// ugly
	sds filename = sdsnew(sourceFile->name);
	filename = sdscat(filename, ".c");
	//strncpy(filename, sourceFile->name, sizeof(char) * (len - 2));
	//filename[len - 2] = '.';
	//filename[len - 1] = 'c';
	//filename[len] = '\0';

	sourceFile->outputFile = fopen(filename, "w");
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
	if (sourceFile) {
		// more ugly pls fix k ty
//		size_t len = strlen(sourceFile->name) + 2;
//		char filename[len + 2];
//		strncpy(filename, sourceFile->name, sizeof(char) * (len - 2));
//		filename[len - 2] = '.';
//		filename[len - 1] = 'c';
//		filename[len] = '\0';
		sds filename = sdsnew(sourceFile->name);
		filename = sdscat(filename, ".c");

		if (!OUTPUT_C) remove(filename);

		destroyHeaderFile(sourceFile->headerFile);
		debugMessage("Destroyed Source File `%s`", sourceFile->name);
		sdsfree(sourceFile->name);
		sdsfree(sourceFile->alloyFileContents);
		free(sourceFile);
	}
}
