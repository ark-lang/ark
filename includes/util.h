#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdlib.h>

#define JAYFOR_VERSION "0.0.0"

/**
 * Colour Printing stuff, not sure
 * if it works on Windows or not
 * so I've disabled it for Windows
 *
 * Update: on windows can confirm, it does not work
 */
#ifdef _WIN32
	#define KNRM  ""
	#define KRED  ""
	#define KGRN  ""
	#define KYEL  ""
	#define KBLU  ""
	#define KMAG  ""
	#define KCYN  ""
	#define KWHT  ""
#elif __linux || __APPLE__
	#define KNRM  "\x1B[0m"
	#define KRED  "\x1B[31m"
	#define KGRN  "\x1B[32m"
	#define KYEL  "\x1B[33m"
	#define KBLU  "\x1B[34m"
	#define KMAG  "\x1B[35m"
	#define KCYN  "\x1B[36m"
	#define KWHT  "\x1B[37m*"
#endif

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

extern void debug_message(char *msg, bool exit_on_error); 

#endif // BOOL_H
