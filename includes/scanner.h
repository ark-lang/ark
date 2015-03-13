#ifndef SCANNER_H
#define SCANNER_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

/**
 * scanner properties,
 * todo: allow multiple files
 * 		 to be scanned.
 */
typedef struct {
	char* contents;
} Scanner;

/**
 * Creates an instance of a scanner
 * @return the scanner
 */
Scanner *createScanner();

/**
 * Reads the given file into
 * `char* contents;`
 */
void scanFile(Scanner *scanner, const char* fileName);

/**
 * Destroys the given scanner
 * @param scanner scanner to destroy
 */
void destroyScanner(Scanner *scanner);

#endif // SCANNER_H