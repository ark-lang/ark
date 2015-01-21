#include "util.h"

char* get_coloured_text(const char *colour, const char *text) {
	// #if __linux || __APPLE__
	// const char *prefix = colour;
	// const char *suffix = "\x1B[00m";
	// size_t alloc_size = sizeof(char) + (strlen(prefix) + strlen(text) + strlen(suffix) + 1);
	// char *result = malloc(alloc_size);
	// result = strcat(result, prefix);
	// result = strcat(result, text);
	// result = strcat(result, suffix);
	// result[alloc_size] = '\0';
	// return result;
	// #else
	// we have to malloc to avoid a seg fault
	// since xyz_message functions free this memory
	char *result = safe_malloc(sizeof(char) * strlen(text));
	result = strcpy(result, text);
	return result;
	// #endif
}

void debug_message(const char *fmt, ...) {
	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		char *temp = get_coloured_text("\x1B[33m", "debug: ");
		fprintf(stdout, "%s", temp);
		free(temp);
		vfprintf(stdout, fmt, arg);
		fprintf(stdout, "\n");
		va_end(arg);
	}
}

void error_message(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = get_coloured_text("\x1B[31m", "error: ");
	fprintf(stdout, "%s", temp);
	free(temp);
	vfprintf(stdout, fmt, arg);
	fprintf(stdout, "\n");
	va_end(arg);
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
