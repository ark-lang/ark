#include "util.h"

char *customStrdup(const char *s) {
	char* data = malloc(strlen(s) + 1);
	if (data)
		strcpy(data, s);
	return data;
}

bool isASCII(char c) {
	// wot
	return ((c & (~0x7f)) == 0);
}

sds randString(size_t length) {
    static char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_";
    sds randomString = sdsempty();

    if (length) {
    	randomString = sdscat(randomString, "__");

        if (randomString) {
            for (size_t n = 2; n < length; n++) {
                int key = rand() % (int) (sizeof(charset) - 1);
                randomString = sdscatlen(randomString, &charset[key], 1);
            }
        }
    }

    return randomString;
}

sds toUppercase(sds str) {
	// TODO: fix this
	size_t len = sdslen(str);
	if (len <= 0) return NULL;

	sds result = sdsnewlen("", len);

	for (size_t i = 0; i < len; i++) {
		result[i] = toupper((unsigned char) str[i]);
	}

	return result;
}

void verboseModeMessage(const char *fmt, ...) {
	if (VERBOSE_MODE) {
		va_list arg;
		va_start(arg, fmt);
		char *temp = GET_ORANGE_TEXT("verbose mode: ");
		fprintf(stdout, "%s", temp);
		vfprintf(stdout, fmt, arg);
		fprintf(stdout, "\n");
		va_end(arg);
	}
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
    if ((retstr = safeMalloc (strlen (file) + 1)) == NULL)
        return NULL;
    strcpy (retstr, file);
    lastdot = strrchr (retstr, '.');
    if (lastdot != NULL)
        *lastdot = '\0';
    return retstr;
}

char *getFileName(char *path) {
	char *s = strrchr(path, '/');
	if (!s) return removeExtension(customStrdup(path));
	char *result = customStrdup(s + 1);
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
				verboseModeMessage("\"%s\" is empty", fileName);
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

void warningMessage(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_ORANGE_TEXT("warning: ");
	fprintf(stderr, "%s", temp);
	vfprintf(stderr, fmt, arg);
	fprintf(stderr, "\n");
	va_end(arg);
}

void warningMessageWithPosition(char *fileName, int lineNumber, int charNumber, const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_ORANGE_TEXT("warning: ");
	fprintf(stderr, "%s", temp);
	if (fileName != NULL)
		fprintf(stderr, "[%s:%d:%d] ", fileName, lineNumber, charNumber);
	else
		fprintf(stderr, "[%d:%d] ", lineNumber, charNumber);
	vfprintf(stderr, fmt, arg);
	fprintf(stderr, "\n");
	va_end(arg);
}

void errorMessage(const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_RED_TEXT("error: ");
	fprintf(stderr, "%s", temp);
	vfprintf(stderr, fmt, arg);
	fprintf(stderr, "\n");
	va_end(arg);
	exit(1);
}

void errorMessageWithPosition(char *fileName, int lineNumber, int charNumber, const char *fmt, ...) {
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_RED_TEXT("error: ");
	fprintf(stderr, "%s", temp);
	if (fileName != NULL)
		fprintf(stderr, "[%s:%d:%d] ", fileName, lineNumber, charNumber);
	else
		fprintf(stderr, "[%d:%d] ", lineNumber, charNumber);
	vfprintf(stderr, fmt, arg);
	fprintf(stderr, "\n");
	va_end(arg);
	exit(1);
}

void errorMessageWithPositionAndLine(char *src, char *fileName, int lineNumber, int lineStart, int charStart, int charEnd, const char *fmt, ...){
	// Print the error message
	va_list arg;
	va_start(arg, fmt);
	char *temp = GET_RED_TEXT("error: ");
	fprintf(stderr, "%s", temp);
	if (fileName != NULL)
		fprintf(stderr, "[%s:%d:%d-%d] ", fileName, lineNumber, charStart, charEnd);
	else
		fprintf(stderr, "[%d:%d-%d] ", lineNumber, charStart, charEnd);
	vfprintf(stderr, fmt, arg);
	fprintf(stderr, "\n");
	va_end(arg);

	// Print the marker and line
	int tokLength = charEnd - charStart;
 	char* line = src + lineStart - charStart;

	for(int i = lineStart - charStart; i < lineStart - charStart + charEnd; line++, i++) fprintf(stderr, "%c", *line);
 	fprintf(stderr, "\n");

 	if(tokLength > 1){
		for(int i = 0; i < charStart - 1; i++) fprintf(stderr, " ");
		fprintf(stderr, "^");
 		for(int i = 0; i < tokLength - 2; i++) fprintf(stderr, "-");
 	}
	fprintf(stderr, "^\n");

	exit(1);
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
	void *memoryChunk = malloc(size);
	assert(memoryChunk);
	return memoryChunk;
}

void *safeCalloc(size_t size) {
	void *memoryChunk = calloc(size, 1);
	assert(memoryChunk);
	return memoryChunk;
}
