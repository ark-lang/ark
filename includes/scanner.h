#ifndef SCANNER_H
#define SCANNER_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

/**
 * Scanner properties,
 * todo: allow multiple files
 * 		 to be scanned.
 */
typedef struct {
	string contents;
} Scanner;

/**
 * Creates an instance of a scanner
 * 
 * @return the scanner
 */
Scanner *scannerCreate();

/**
 * Reads the given file into
 * `string contents;`
 * 
 */
void scannerReadFile(Scanner *scanner, const string fileName);

/**
 * Destroys the given Scanner
 * 
 * @param scanner scanner to destroy
 */
void scannerDestroy(Scanner *scanner);

#endif // SCANNER_H