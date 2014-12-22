#ifndef BOOL_H
#define BOOL_H

#define jayfor_VERSION "0.0.0"

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

extern bool DEBUG_MODE;
extern bool EXECUTE_BYTECODE;
extern char* EXECUTABLE_FILENAME;

#endif // BOOL_H
