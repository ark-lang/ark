#ifndef __HEADER_FILE_H
#define __HEADER_FILE_H

#include <stdlib.h>

#include "util.h"

/**
 * Contains the properties of a Headerfile
 */
typedef struct {
	/** name of the header file */
	sds fileName;

	/** name of the header file, with the extension, path, etc... */
	char* name;

	/** the generated file name with a _gen_ prefix and .h extension */
	sds generatedHeaderName;

	/** the file to output */
	FILE *outputFile;
} HeaderFile;

/**
 * Create a HeaderFile
 *
 * @param fileName the path to the file
 */
HeaderFile *createHeaderFile(char *fileName);

/**
 * Write the header file
 */
void writeHeaderFile(HeaderFile *headerFile);

/**
 * Close the header file and free up any unused
 * things
 */
void closeHeaderFile(HeaderFile *headerFile);

/**
 * Destroy the header file, clean up after everything
 */
void destroyHeaderFile(HeaderFile *headerFile);

#endif // __HEADER_FILE_H
