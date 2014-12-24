#ifndef BOOL_H
#define BOOL_H

#include <stdio.h>
#include <stdlib.h>

/** the current version of jayfor */
#define JAYFOR_VERSION "0.0.0"

#ifdef _WIN32
	#define KNRM()  ""
	#define KRED()  ""
	#define KGRN()  ""
	#define KYEL()  ""
	#define KBLU()  ""
	#define KMAG()  ""
	#define KCYN()  ""
	#define KWHT()  ""
#else
	#define KNRM()  "printf(\x1B[0m);"
	#define KRED()  "printf(\x1B[31m);"
	#define KGRN()  "printf(\x1B[32m);"
	#define KYEL()  "printf(\x1B[33m);"
	#define KBLU()  "printf(\x1B[34m);"
	#define KMAG()  "printf(\x1B[35m);"
	#define KCYN()  "printf(\x1B[36m);"
	#define KWHT()  "printf(\x1B[37m*);"
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

/**
 * Emitts a debug message to the console if we are in DEBUG MODE
 * @param msg           the message to print
 * @param exit_on_error if we should exit on an error
 */
extern void debug_message(const char *fmt, ...); 

extern void error_message(const char *fmt, ...);

extern void primary_message(const char *fmt, ...);

#endif // BOOL_H
