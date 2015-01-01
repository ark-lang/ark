#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/** the current version of jayfor */
#define JAYFOR_VERSION "0.0.0"

/** quick boolean implementation */
typedef enum {
	false, true
} bool;

/** if we are in debug mode -- will print debug warnings */
extern bool DEBUG_MODE;

/** if we are going to execute the bytecode after generation */
extern bool EXECUTE_BYTECODE;

/** if we are running a vm executable */
extern bool RUN_VM_EXECUTABLE;

/** the name of vm executable name, if applicable */
extern char* VM_EXECUTABLE_NAME;

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
