#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

/** the current version of jayfor */
#define JAYFOR_VERSION "0.0.0"

/** if we are in debug mode -- will print debug warnings */
extern bool DEBUG_MODE;

/** the name of the executable file */
extern char* OUTPUT_EXECUTABLE_NAME;

/**
 * Emitts a debug message to the console if we are in DEBUG MODE
 * @param msg           the message to print
 * @param ...			extra arguments
 */
extern void debug_message(const char *fmt, ...); 

/**
 * Emitts an error message to the console, will also exit
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
extern void error_message(const char *fmt, ...);

/**
 * Emitts a primary message to the console
 * @param msg           the message to print
 * @param ... 			extra arguments
 */
extern void primary_message(const char *fmt, ...);

/**
 * Gets the extension of the given file
 * @param  filename the filename to get the extension
 * @return          the extension of the file given
 */
extern const char *get_filename_ext(const char *filename);

#endif // BOOL_H
