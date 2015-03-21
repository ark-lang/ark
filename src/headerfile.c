#include "headerfile.h"

HeaderFile *createHeaderFile(char *fileName) {
	HeaderFile *headerFile = safeMalloc(sizeof(*headerFile));
	headerFile->fileName = fileName;
	headerFile->name = getFileName(headerFile->fileName);
	return headerFile;
}

void writeHeaderFile(HeaderFile *headerFile) {
	// ugly
	size_t len = strlen(headerFile->name) + 2;
	char filename[len + 2];
	strncpy(filename, headerFile->name, sizeof(char) * (len - 2));
	filename[len - 2] = '.';
	filename[len - 1] = 'h';
	filename[len] = '\0';

	headerFile->outputFile = fopen(filename, "w");
	if (!headerFile->outputFile) {
		perror("fopen: failed to open file");
		return;
	}
}

void closeHeaderFile(HeaderFile *headerFile) {
	fclose(headerFile->outputFile);
}

void destroyHeaderFile(HeaderFile *headerFile) {
	if (headerFile) {
		// more ugly pls fix k ty
		size_t len = strlen(headerFile->name) + 2;
		char filename[len + 2];
		strncpy(filename, headerFile->name, sizeof(char) * (len - 2));
		filename[len - 2] = '.';
		filename[len - 1] = 'h';
		filename[len] = '\0';
		if (!OUTPUT_C) remove(filename);

		debugMessage("Destroyed Header File `%s`", headerFile->name);
		free(headerFile->name);
		free(headerFile);
	}
}
