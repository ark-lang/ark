#include "scanner.h"

scanner *create_scanner() {
	scanner *self = malloc(sizeof(*self));
	if (!self) {
		perror("malloc: failed to allocate memory for scanner");
		exit(1);
	}
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
				exit(1);
			}

			self->contents = malloc(sizeof(*self->contents) * (fileSize + 1));
			if (!self->contents) {
				perror("malloc: failed to allocate memory for file");
				exit(1);
			}

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				exit(1);
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
		exit(1);
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