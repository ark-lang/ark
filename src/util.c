#include "util.h"

void debug_message(const char *fmt, ...) {
	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		#if __linux || __APPLE__
		fprintf(stdout, "\x1B[31m");
		#endif
		vfprintf(stdout, fmt, arg);
		#if __linux || __APPLE__
		fprintf(stdout, "\x1B[0m");
		#endif
		fprintf(stdout, "\n");
		va_end(arg);
	}
}

void error_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	#if __linux || __APPLE__
	fprintf(stdout, "\x1B[33m");
	#endif
	vfprintf(stdout, fmt, arg);
	#if __linux || __APPLE__
	fprintf(stdout, "\x1B[0m");
	#endif
	fprintf(stdout, "\n");
	va_end(arg);
	exit(1);
}

void primary_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	vfprintf(stdout, fmt, arg);
	fprintf(stdout, "\n");
	va_end(arg);
} 

const char *get_filename_ext(const char *filename) {
	const char *dot = strrchr(filename, '.');
	if (!dot || dot == filename) return "";
	return dot + 1;
}

void *safe_malloc(size_t size) {
	void *ret = malloc(size);
	if (!ret) {
		fprintf(stderr, "malloc: failed to allocate %ld bytes of memory: %s\n", (long)size, strerror(errno));
		exit(1);
	}
	return ret;
}
