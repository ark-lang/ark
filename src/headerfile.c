#include "headerfile.h"

HeaderFile *createHeaderFile(sds fileName) {
	HeaderFile *headerFile = safeMalloc(sizeof(*headerFile));
	headerFile->fileName = fileName;
	headerFile->name = getFileName(headerFile->fileName);
	return headerFile;
}

void writeHeaderFile(HeaderFile *headerFile) {
	// ugly
//	size_t len = strlen(headerFile->name) + 2;
//	char filename[len + 2];
//	strncpy(filename, headerFile->name, sizeof(char) * (len - 2));
//	filename[len - 2] = '.';
//	filename[len - 1] = 'h';
//	filename[len] = '\0';
	sds filename = sdsnew(headerFile->name);
	filename = sdscat(filename, ".h");

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
//		size_t len = strlen(headerFile->name) + 2;
//		char filename[len + 2];
//		strncpy(filename, headerFile->name, sizeof(char) * (len - 2));
//		filename[len - 2] = '.';
//		filename[len - 1] = 'h';
//		filename[len] = '\0';
		sds filename = sdsnew(headerFile->name);
		filename = sdscat(filename, ".h");

		if (!OUTPUT_C) remove(filename);

		sdsfree(filename);
		debugMessage("Destroyed Header File `%s`", headerFile->name);
		sdsfree(headerFile->name);
		free(headerFile);
	}
}
