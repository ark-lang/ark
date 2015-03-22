#include "headerfile.h"

HeaderFile *createHeaderFile(sds fileName) {
	HeaderFile *headerFile = safeMalloc(sizeof(*headerFile));
	headerFile->fileName = fileName;
	headerFile->name = getFileName(headerFile->fileName);
	return headerFile;
}

void writeHeaderFile(HeaderFile *headerFile) {
	headerFile->generatedHeaderName = sdsempty();
	sdscat(headerFile->generatedHeaderName, "_gen_");
	sdscat(headerFile->generatedHeaderName, headerFile->name);
	sdscat(headerFile->generatedHeaderName, ".h");

	debugMessage("Generated header file `%s` as `%s`", headerFile->name, headerFile->generatedHeaderName);
	headerFile->outputFile = fopen(headerFile->generatedHeaderName, "w");
	if (!headerFile->outputFile) {
		perror("fopen: failed to open file");
		return;
	}
}

void closeHeaderFile(HeaderFile *headerFile) {
	fclose(headerFile->outputFile);
}

void destroyHeaderFile(HeaderFile *headerFile) {
	if (!OUTPUT_C) remove(headerFile->generatedHeaderName);

	debugMessage("Destroyed Header File `%s`", headerFile->name);
	sdsfree(headerFile->generatedHeaderName);
	free(headerFile->name);
	free(headerFile);
}
