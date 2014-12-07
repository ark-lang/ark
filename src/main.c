#include <stdio.h>
#include <stdlib.h>

#include "lexer.h"

char *readFile(const char *fileName) {
	FILE *file = fopen(fileName, "r");
	char *contents = NULL;

	if (file) {
		if (!fseek(file, 0, SEEK_END)) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				exit(1);
			}

			contents = malloc(sizeof(*contents) * (fileSize + 1));
			if (!contents) {
				perror("malloc: failed to allocate memory for file");
				exit(1);
			}

			if (fseek(file, 0, SEEK_SET)) {
				perror("could not reset file index");
				exit(1);
			}

			size_t fileLength = fread(contents, sizeof(char), fileSize, file);
			if (!fileLength) {
				perror("fread: file is empty");
				exit(1);
			}

			contents[fileSize] = '\0';
		}
		fclose(file);
	}
	else {
		perror("fopen: could not read file");
		exit(1);
	}

	return contents;
}

void startCompiling(char *source) {
	Lexer *lexer = lexerCreate(source);

	do {
		lexerGetNextToken(lexer);
		printf("%s\n", lexer->token.content);
	}
	while (lexer->token.type != END_OF_FILE);

	lexerDestroy(lexer);
}

int main(int argc, char** argv) {
	if (argc > 1) {
		char *fileName = argv[1];
		char *sourceCode = readFile(fileName);
		startCompiling(sourceCode);
		free(sourceCode);
	}
	else {
		printf("error: no input files\n");
		exit(1);
	}

	return 0;
}
