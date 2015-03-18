#include "headerfile.h"

HeaderFile *createHeaderFile(char *fileName) {
	HeaderFile *headerFile = malloc(sizeof(*headerFile));
	headerFile->fileName = fileName;
	headerFile->name = getFileName(headerFile->fileName);
	return headerFile;
}

void writeHeaderFile(HeaderFile *headerFile) {
	headerFile->outputFile = fopen(headerFile->name ".h", "w");
	if (!headerFile->outputFile) {
		perror("fopen: failed to open file");
		return;
	}
}

void closeHeaderFile(HeaderFile *headerFile) {
	fclose(headerFile->outputFile);
	free(headerFile->name);
}

void destroyHeaderFile(HeaderFile *headerFile) {
	if (headerFile) {
		free(headerFile);
	}
}
