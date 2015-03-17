#ifndef HEADER_FILE_H
#define HEADER_FILE_H

#include <stdlib.h>

#include "util.h"

typedef struct {
	char *fileName;
	char *fileContents;

	char *outputFileContents;
} HeaderFile;

HeaderFile *createHeaderFile(char *fileName);

void writeHeaderFile(HeaderFile *headerFile);

void destroyHeaderFile(HeaderFile *headerFile);

#endif // HEADER_FILE_H
