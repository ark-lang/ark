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
	sourceFile->generatedSourceName = sdsempty();
	sdscat(sourceFile->generatedSourceName, "__gen_");
	sdscat(sourceFile->generatedSourceName, sourceFile->name);
	sdscat(sourceFile->generatedSourceName, ".c");

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
	if (sourceFile) {
		if (!OUTPUT_C) remove(sourceFile->generatedSourceName);

		destroyHeaderFile(sourceFile->headerFile);

		debugMessage("Destroyed Source File `%s`", sourceFile->name);
		sdsfree(sourceFile->fileName);
		free(sourceFile->name); // this isn't using sds!
		sdsfree(sourceFile->generatedSourceName);
		free(sourceFile->alloyFileContents); // and this isn't either! vedant u fkn noob
		free(sourceFile);
	}
}
