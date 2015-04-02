#ifndef SOURCE_FILE_H
#define SOURCE_FILE_H

#include <stdlib.h>

#include "vector.h"
#include "headerfile.h"
#include "util.h"

/**
 * SourceFile properties
 */
typedef struct {
	sds fileName;				// file name for the source file
	char* name;					// file name for the source file, raw
	char* alloyFileContents;	// the contents of the alloy file
	sds generatedSourceName;	// the generated source name with the _gen_ prefix and .c suffix
	FILE *outputFile;			// the file output

	Vector *tokens;				// the token stream for the source file
	Vector *ast;				// the output AST tree

	HeaderFile *headerFile;		// the header file for the source file
} SourceFile;

/**
 * Create a source file with the given file
 *
 * @param fileName the alloy file
 */
SourceFile *createSourceFile(sds fileName);

/**
 * Write both the header and source files
 *
 * @param sourceFile the sourceFile instance
 */
void writeFiles(SourceFile *sourceFile);

/**
 * Write the source file
 *
 * @param sourceFile the sourceFile to write
 */
void writeSourceFile(SourceFile *sourceFile);

/**
 * Close the given sourceFile
 *
 * @param sourceFile the sourceFile to close
 */
void closeSourceFile(SourceFile *sourceFile);

/**
 * Close the header and source files
 *
 * @param sourceFile the sourceFile instance,
 * 		  which gives access to the sourceFile & headerFile
 */
void closeFiles(SourceFile *sourceFile);

/**
 * Destroy the source file
 *
 * @param sourceFile the sourceFile to destroy
 */
void destroySourceFile(SourceFile *sourceFile);

#endif // SOURCE_FILE_H