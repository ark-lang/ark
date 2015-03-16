#include "scanner.h"

Scanner *createScanner() {
	Scanner *self = safeMalloc(sizeof(*self));
	self->contents = NULL;
	return self;
}

void scanFile(Scanner *self, const char* fileName) {
	FILE *file = fopen(fileName, "rb");

	if (file) {
		if (!fseek(file, 0, SEEK_END)) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				return;
			}

			self->contents = safeMalloc(sizeof(*self->contents) * (fileSize + 1));

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				return;
			}

			size_t fileLength = fread(self->contents, sizeof(char), fileSize, file);
			if (!fileLength) {
				debugMessage("warning: \"%s\" is empty\n", fileName);
			}

			self->contents[fileSize] = '\0';
		}
		fclose(file);
	}
	else {
		perror("fopen: could not read file");
		printf("file: %s\n", fileName);
		return;
	}
}

void destroyScanner(Scanner *self) {
	if (self) {
		if (self->contents) {
			free(self->contents);
		}
		free(self);
	}
}
