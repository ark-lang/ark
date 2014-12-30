#include "util.h"

extern void debug_message(const char *fmt, ...) {
	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		KRED();
		vfprintf(stdout, fmt, arg);
		KNRM();
		va_end(arg);
	}
}

extern void error_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	KYEL();
	vfprintf(stdout, fmt, arg);
	KNRM();
	va_end(arg);
	exit(1);
}

extern void primary_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	KGRN();
	vfprintf(stdout, fmt, arg);
	KNRM();
	va_end(arg);
} 

extern const char *get_filename_ext(const char *filename) {
	const char *dot = strrchr(filename, '.');
	if (!dot || dot == filename) return "";
	return dot + 1;
}