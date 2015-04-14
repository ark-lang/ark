#ifndef __BOOL_H
#define __BOOL_H

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <errno.h>
#include <ctype.h>
#include <assert.h>
#include <time.h>

#include "sds.h"

/** the current version of alloy */
#define ALLOYC_VERSION "0.0.7"
#define ALLOY_UNUSED_OBJ(x) (void)(x)

/** windows doesn't like coloured text */
#ifdef _WIN32
	#define GET_RED_TEXT(x) (x)
	#define GET_ORANGE_TEXT(x) (x)
#else
	#define GET_RED_TEXT(x) ("\x1B[31m" x "\x1B[00m")
	#define GET_ORANGE_TEXT(x) ("\x1B[33m" x "\x1B[00m")
#endif

/** macro for array length */
#define ARR_LEN(x) (int) (sizeof(x) / sizeof(x[0]))

/** flags */
extern bool DEBUG_MODE;
extern bool VERBOSE_MODE;
extern char* OUTPUT_EXECUTABLE_NAME;
extern bool OUTPUT_C;
extern char* ADDITIONAL_COMPILER_ARGS;
extern char* COMPILER;

/**
 * Returns if the char is ASCII
 * @param  c the char to check
 * @return   true if the char is ascii
 */
bool isASCII(char c);

/**
 * Strdup so we can keep everything C11 compliant
 */
char* alloyStrdup(const char* s);

/**
 * Generates a random string of the given length
 *
 * @param length the length of the string
 */
char *randString(size_t length);

/**
 * Converts a string to uppercase
 *
 * @param str the string to convert to uppercase
 */
char *toUppercase(char *str);

/**
 * Removes the extension from the given file
 */
char *removeExtension(char *file);

/**
 * Cut the directories from a path
 */
char *getFileName(char *path);

/**
 * Emits a debug message to the console if we are in DEBUG MODE
 * @param msg           the message to print
 * @param ...			extra arguments
 */
void verboseModeMessage(const char *fmt, ...);

/**
 * Emits a warning message to the console
 * @param msg           the message to print
 * @param ...			extra arguments
 */
void warningMessage(const char *fmt, ...);

/**
 * Emits a message when in verbose mode
 * @param msg           the message to print
 * @param ...			extra arguments
 */
void verboseModeMessage(const char *fmt, ...);

/**
 * Emits an error message to the console, will also exit
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void errorMessage(const char *fmt, ...);

/**
 * Emits a primary message to the console
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

#endif // __BOOL_H
