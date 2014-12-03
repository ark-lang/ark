#include <stdio.h>
#include <stdlib.h>

#include "lexer.h"

char *readFile(const char *fileName) {
	FILE *file = fopen(fileName, "r");
	char *contents;

	if (file) {
		if (fseek(file, 0, SEEK_END)) {
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

int main(int argc, char** argv) {
	if (argc > 0) {
		int i;
		for (i = 0; i < argc; i++) {
			printf("arg: %s\n", argv[i]);
		}
	}

	Lexer *lexer = createLexer("this is a test 123 wow");

	do {
		getNextToken(lexer);
		switch (lexer->token.class) {
			case IDENTIFIER:	printf("identifier"); 	break;
			case INTEGER:		printf("integer"); 		break;
			case ERRORNEOUS:	printf("error");		break;
			case END_OF_FILE:	printf("end of file");	break;
			default:			printf("op or sep");  	break;
		}
		printf(": %s\n", lexer->token.repr);
	}
	while (lexer->token.class != END_OF_FILE);

	destroyLexer(lexer);

	return 0;
}