#ifndef SOURCE_FILE_H
#define SOURCE_FILE_H

#include <stdlib.h>

#include "vector.h"
#include "headerfile.h"
#include "util.h"

typedef struct {
	char *fileName;
	char *name;
	char *alloyFileContents;
	FILE *outputFile;

	Vector *tokens;
	Vector *ast;

	HeaderFile *headerFile;
} SourceFile;

SourceFile *createSourceFile(char *fileName);

void writeFiles(SourceFile *sourceFile);

void writeSourceFile(SourceFile *sourceFile);

void closeSourceFile(SourceFile *sourceFile);

void closeFiles(SourceFile *sourceFile);

void destroySourceFile(SourceFile *sourceFile);

#endif // SOURCE_FILE_H
