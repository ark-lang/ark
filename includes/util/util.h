#ifndef __UTIL_H
#define __UTIL_H

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

/** the name and version of the compiler */
#define COMPILER_NAME          "Ark"
#define COMPILER_VERSION       "0.0.8"

#define COMPILER_UNUSED_OBJ(x) (void)(x)

#if defined(_WIN32) || defined(__WIN32__) || defined (WIN32)
	#define WINDOWS
#endif

/** windows doesn't like coloured text */
#ifdef WINDOWS
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
extern char *OUTPUT_EXECUTABLE_NAME;
extern bool OUTPUT_C;
extern char *ADDITIONAL_COMPILER_ARGS;
extern char *COMPILER;
extern char *linkerFlags;
extern bool IGNORE_MAIN;

/**
 * Returns if the char is ASCII
 * @param  c the char to check
 * @return   true if the char is ascii
 */
bool isASCII(char c);

/**
 * Custom strdup so we can keep everything C11 compliant
 * @param s the char to dup
 * @return  a new char instance
 */
char *customStrdup(const char *s);

/**
 * Generates a random string of the given length
 * @param length the length of the string
 */
char *randString(size_t length);

/**
 * Converts a string to uppercase
 * @param str the string to convert to uppercase
 */
char *toUppercase(char *str);

/**
 * Removes the extension from the given file
 * @param file the file to remove the extension from
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
 * Emits a warning message to the console
 * The warning message contains a line and character number.
 * @param fileName      the filename to print (can be NULL)
 * @param lineNumber    the line number to print
 * @param charNumber    the character number to print
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void warningMessageWithPosition(char *fileName, int lineNumber, int charNumber, const char *fmt, ...);

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
 * Emits an error message to the console, will also exit
 * The error message contains a line and character number.
 * @param fileName      the filename to print (can be NULL)
 * @param lineNumber    the line number to print
 * @param charNumber    the character number to print
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void errorMessageWithPosition(char *fileName, int lineNumber, int charNumber, const char *fmt, ...);

/**
 * Emits an error message to the console, will also exit
 * The error message contains a line and character number.
 * After emitting the error message, the errornous is displayed.
 * @param fileName      the filename to print (can be NULL)
 * @param lineNumber    the line number to print
 * @param charNumber    the character number to print
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void errorMessageWithPositionAndLine(char* src, char *fileName, int lineNumber, int lineStart, int charNumber, int charEnd, const char *fmt, ...);

/**
 * Emits a primary message to the console
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
void primaryMessage(const char *fmt, ...);

/**
 * Reads a file, returns the contents of the file
 * or NULL if it failed.
 */
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

/**
 * Safe calloc, dies if allocation fails
 * @param  size size of space to allocate
 * @return pointer to allocated data
 */
void *safeCalloc(size_t size);

#endif // __UTIL_H
