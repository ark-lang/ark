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
	char* contents;
} Scanner;

/**
 * Creates an instance of a scanner
 * @return the scanner
 */
Scanner *create_scanner();

/**
 * Reads the given file into
 * `char* contents;`
 */
void scan_file(Scanner *scanner, const char* fileName);

/**
 * Destroys the given Scanner
 * @param scanner scanner to destroy
 */
void destroy_scanner(Scanner *scanner);

#endif // SCANNER_H