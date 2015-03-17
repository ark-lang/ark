#include "headerfile.h"

HeaderFile *createHeaderFile(char *fileName) {
	HeaderFile *headerFile = malloc(sizeof(*headerFile));
	headerFile->fileName = fileName;
	headerFile->fileContents = malloc(sizeof(char));
	headerFile->fileContents[0] = '\0';

	strcpy(headerFile->outputFileName, fileName);
	str_append(headerFile->outputFileName, ".h");
	printf("%s\n", headerFile->outputFileName);

	return headerFile;
}

void writeHeaderFile(HeaderFile *headerFile) {
	FILE *file = fopen(headerFile->outputFileName, "w");
	if (!file) {
		perror("fopen: failed to open file");
		return;
	}

	fprintf(file, "%s", headerFile->fileContents);
	fclose(file);

	system(JOIN_STR(COMPILER, headerFile->outputFileName));
}

void destroyHeaderFile(HeaderFile *headerFile) {
	if (headerFile) {
		remove(headerFile->outputFileName);
		free(headerFile->fileContents);
		free(headerFile);
	}
}