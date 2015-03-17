#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <errno.h>
#include <assert.h>
#include <time.h>

/** the current version of alloy */
#define ALLOYC_VERSION "0.0.2"

#define JOIN_STR(x, y) #x " " #y

/** windows doesn't like coloured text */
#ifdef _WIN32
	#define GET_RED_TEXT(x) (x)
	#define GET_ORANGE_TEXT(x) (x)
#else
	#define GET_RED_TEXT(x) ("\x1B[31m" x "\x1B[00m")
	#define GET_ORANGE_TEXT(x) ("\x1B[33m" x "\x1B[00m")
#endif

/** what compiler to compile generated code in, default is GCC for now */
#define COMPILER "gcc"

/** if we are in debug mode -- will print debug warnings */
extern bool DEBUG_MODE;

/** the name of the executable file */
extern char* OUTPUT_EXECUTABLE_NAME;

void str_append(char *original_str, char *str);

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
