#include "util.h"

extern void debug_message(const char *fmt, ...) {
	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		KYEL();
		printf(fmt, arg);
		fflush(stdout);
		KNRM();
		va_end(arg);
	}
}

extern void error_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	KYEL();
	printf(fmt, arg);
	fflush(stdout);
	KNRM();
	va_end(arg);
	exit(1);
}

extern void primary_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	KYEL();
	printf(fmt, arg);
	KNRM();
	va_end(arg);
	exit(1);
}