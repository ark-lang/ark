#ifndef PREPROCESSOR_H
#define PREPROCESSOR_H

/**
 * The preprocessor will take the contents of a file and scan
 * for preprocessor directives process them then remove them from
 * the contents of the file to be passed to the lexer.
 */

#include "util.h"

#include <stdlib.h>

typedef struct {
	char *file_content;
	char *file_processed;
} preprocessor;

preprocessor *create_preprocessor(char *file_content);

char *start_preprocessing(preprocessor *self);

void destroy_preprocessor(preprocessor *self);

#endif // PREPROCESSOR_H