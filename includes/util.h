#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <errno.h>
#include <ctype.h>
#include <assert.h>
#include <time.h>

/** the current version of alloy */
#define ALLOYC_VERSION "0.0.3"

/** windows doesn't like coloured text */
#ifdef _WIN32
	#define GET_RED_TEXT(x) (x)
	#define GET_ORANGE_TEXT(x) (x)
#else
	#define GET_RED_TEXT(x) ("\x1B[31m" x "\x1B[00m")
	#define GET_ORANGE_TEXT(x) ("\x1B[33m" x "\x1B[00m")
#endif

extern bool DEBUG_MODE;
extern char* COMPILER;
extern bool OUTPUT_C;
extern char* OUTPUT_EXECUTABLE_NAME;

char *randString(size_t length);

char *toUppercase(char *str);

char *joinString(const char* s1, const char* s2);

char *removeExtension(char *file);

char *getFileName(char *path);

void appendString(char *original_str, char *str);

/**
 * Emitts a debug message to the console if we are in DEBUG MODE
 * @param msg           the message to print
 * @param ...			extra arguments
 */
void debugMessage(const char *fmt, ...);

/**
 * Emitts an error message to the console, will also exit
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void errorMessage(const char *fmt, ...);

/**
 * Emitts a primary message to the console
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void primaryMessage(const char *fmt, ...);

char *readFile(const char *fileName);

/**
 * Gets the extension of the given file
 * @param  filename the filename to get the extension
 * @return          the extension of the file given
 */
const char *getFilenameExtension(const char *filename);

/**
 * Safe malloc, dies if allocation fails
 * @param  size size of space to allocate
 * @return pointer to allocated data
 */
void *safeMalloc(size_t size);

#endif // BOOL_H
