#include "scanner.h"

scanner* create_scanner() {
	scanner *self = safe_malloc(sizeof(*self));
	self->contents = NULL;
	return self;
}

void scan_file(scanner *self, const char* fileName) {
	FILE *file = fopen(fileName, "rb");

	if (file) {
		if (!fseek(file, 0, SEEK_END)) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				return;
			}

			self->contents = safe_malloc(sizeof(*self->contents) * (fileSize + 1));

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				return;
			}

			size_t fileLength = fread(self->contents, sizeof(char), fileSize, file);
			if (!fileLength) {
				debug_message("warning: \"%s\" is empty\n", fileName);
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

void destroy_scanner(scanner *self) {
	if (self) {
		if (self->contents) {
			free(self->contents);
		}
		free(self);
	}
}
