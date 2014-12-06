#include <stdio.h>
#include <stdlib.h>

#include "lexer.h"

char *readFile(const char *fileName) {
	FILE *file = fopen(fileName, "r");
	char *contents = NULL;

	if (file != NULL) {
		if (fseek(file, 0, SEEK_END) == 0) {
			long fileSize = ftell(file);
			if (fileSize == -1) {
				perror("ftell: could not read filesize");
				exit(1);
			}
			contents = malloc(sizeof(*contents) * (fileSize + 1));
			if (contents == NULL) {
				perror("malloc: failed to allocate memory for file");
				exit(1);
			}

			if (fseek(file, 0, SEEK_SET) != 0) {
				perror("could not reset file index");
				exit(1);
			}

			size_t fileLength = fread(contents, sizeof(char), fileSize, file);
			if (fileLength == 0) {
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
	Lexer *lexer = createLexer(source);
	// only a 10th of a megabyte, fuck it
	// todo do this dynamically
	Token tokens[TOKEN_LIST_MAX_SIZE];

	int index = 0;
	do {
		getNextToken(lexer);
		if (index > TOKEN_LIST_MAX_SIZE) {
			printf("token overflow!\n");
			exit(1);
		}
		printf("%s\n", lexer->token.repr);
		tokens[index++] = lexer->token;
	}
	while (lexer->token.class != END_OF_FILE);

	destroyLexer(lexer);
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
