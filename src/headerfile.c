#include "headerfile.h"

HeaderFile *createHeaderFile(char *fileName) {
	HeaderFile *headerFile = malloc(sizeof(*headerFile));
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
	printf("%s\n", filename);

	headerFile->outputFile = fopen(filename, "w");
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
