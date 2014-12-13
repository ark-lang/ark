#include "scanner.h"

Scanner *scannerCreate() {
	Scanner *scanner = malloc(sizeof(*scanner));
	if (!scanner) {
		perror("malloc: failed to allocate memory for scanner");
		exit(1);
	}
	scanner->contents = NULL;
	return scanner;
}

void scannerReadFile(Scanner *scanner, const char* fileName) {
	FILE *file = fopen(fileName, "rb");

	if (file) {
		if (!fseek(file, 0, SEEK_END)) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				exit(1);
			}

			scanner->contents = malloc(sizeof(*scanner->contents) * (fileSize + 1));
			if (!scanner->contents) {
				perror("malloc: failed to allocate memory for file");
				exit(1);
			}

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				exit(1);
			}

			size_t fileLength = fread(scanner->contents, sizeof(char), fileSize, file);
			if (!fileLength) {
				printf(KYEL "warning: \"%s\" is empty\n" KNRM, fileName);
			}

			scanner->contents[fileSize] = '\0';
		}
		fclose(file);
	}
	else {
		perror("fopen: could not read file");
		exit(1);
	}
}

void scannerDestroy(Scanner *scanner) {
	free(scanner->contents);
	free(scanner);
}