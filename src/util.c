#include "util.h"

void str_append(char *original_str, char *str) {
	original_str = realloc(original_str, strlen(original_str) + strlen(str) + 1);
	strcat(original_str, str);
}

void debugMessage(const char *fmt, ...) {
	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		char *temp = GET_ORANGE_TEXT("debug: ");
		fprintf(stdout, "%s", temp);
		vfprintf(stdout, fmt, arg);
		fprintf(stdout, "\n");
		va_end(arg);
	}
}

char *removeExtension(char *file) {
    char *retstr;
    char *lastdot;
    if (file == NULL)
         return NULL;
    if ((retstr = malloc (strlen (file) + 1)) == NULL)
        return NULL;
    strcpy (retstr, file);
    lastdot = strrchr (retstr, '.');
    if (lastdot != NULL)
        *lastdot = '\0';
    return retstr;
}

char *getFileName(char *path) {
	char *s = strrchr(path, '/');
	if (!s) return removeExtension(strdup(path));
	char *result = strdup(s + 1);
	char *resultWithoutExt = removeExtension(result);
	return resultWithoutExt;
}

char *readFile(const char *fileName) {
	FILE *file = fopen(fileName, "r");
	char *contents = NULL;

	if (file) {
		if (!fseek(file, 0, SEEK_END)) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				return NULL;
			}

			contents = safeMalloc(sizeof(*contents) * (fileSize + 1));

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				return NULL;
			}

			size_t fileLength = fread(contents, sizeof(char), fileSize, file);
			if (!fileLength) {
				debugMessage("warning: \"%s\" is empty\n", fileName);
			}

			contents[fileLength] = '\0';
		}
		fclose(file);
		return contents;
	}
	else {
		perror("fopen: could not read file");
		return NULL;
	}

	return NULL;
}

void errorMessage(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_RED_TEXT("error: ");
	fprintf(stdout, "%s", temp);
	vfprintf(stdout, fmt, arg);
	fprintf(stdout, "\n");
	va_end(arg);
}

void primaryMessage(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	vfprintf(stdout, fmt, arg);
	fprintf(stdout, "\n");
	va_end(arg);
} 

const char *getFilenameExtension(const char *filename) {
	const char *dot = strrchr(filename, '.');
	if (!dot || dot == filename) return "";
	return dot + 1;
}

void *safeMalloc(size_t size) {
	void *mem_chunk = malloc(size);
	assert(mem_chunk);
	return mem_chunk;
}
