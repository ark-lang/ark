#include <stdio.h>

#include "lexer.h"

int main(int argc, char** argv) {
	Lexer *lexer = createLexer("this is a test 123 wow");

	do {
		getNextToken(lexer);
		switch (lexer->token.class) {
			case IDENTIFIER:	printf("identifier"); break;
			case INTEGER:		printf("integer"); 	break;
			case ERRORNEOUS:	printf("error");		break;
			case END_OF_FILE:	printf("end of file");	break;
			default:			printf("op or sep");  break;
		}
		printf(": %s\n", lexer->token.repr);
	}
	while (lexer->token.class != END_OF_FILE);

	destroyLexer(lexer);

	return 0;
}