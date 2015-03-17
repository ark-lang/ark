#ifndef SOURCE_FILE_H
#define SOURCE_FILE_H

#include <stdlib.h>

#include "vector.h"
#include "headerfile.h"
#include "util.h"

typedef struct {
	char *fileName;
	char *fileContents;
	char *outputFileName;
	Vector *nodes;

	HeaderFile *headerFile;
} SourceFile;

SourceFile *createSourceFile(char *fileName, Vector *nodes);

void writeFiles(SourceFile *sourceFile);

void writeSourceFile(SourceFile *sourceFile);

void destroySourceFile(SourceFile *sourceFile);

#endif // SOURCE_FILE_H