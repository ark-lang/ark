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
} scanner;

/**
 * Creates an instance of a scanner
 * @return the scanner
 */
scanner* create_scanner();

/**
 * Reads the given file into
 * `char* contents;`
 */
void scan_file(scanner *scanner, const char* fileName);

/**
 * Destroys the given scanner
 * @param scanner scanner to destroy
 */
void destroy_scanner(scanner *scanner);

#endif // SCANNER_H
