#ifndef SCANNER_H
#define SCANNER_H

#include <stdio.h>
#include <stdlib.h>

#include "util.h"

typedef struct {
	string contents;
} Scanner;

Scanner *scannerCreate();

void scannerReadFile(Scanner *scanner, const string fileName);

void scannerDestroy(Scanner *scanner);

#endif // SCANNER_H