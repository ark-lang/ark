#ifndef SOURCE_FILE_H
#define SOURCE_FILE_H

#include <stdlib.h>

#include "vector.h"
#include "util.h"

/**
 * SourceFile properties
 */
typedef struct {
	sds fileName;				// file name for the source file
	char* name;					// file name for the source file, raw
	char* alloyFileContents;	// the contents of the alloy file
	FILE *outputFile;			// the file output

	Vector *tokens;				// the token stream for the source file
	Vector *ast;				// the output AST tree
} SourceFile;

/**
 * Create a source file with the given file
 *
 * @param fileName the alloy file
 */
SourceFile *createSourceFile(sds fileName);

/**
 * Destroy the source file
 *
 * @param sourceFile the sourceFile to destroy
 */
void destroySourceFile(SourceFile *sourceFile);

#endif // SOURCE_FILE_H
