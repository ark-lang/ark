#include "jayfor.h"

Jayfor *jayforCreate(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	}

	// create the instance of jayfor
	Jayfor *jayfor = malloc(sizeof(*jayfor));
	if (!jayfor) {
		perror("malloc: failed to allocate memory for JAYFOR");
		exit(1);
	}

	// just in case.
	jayfor->scanner = NULL;
	jayfor->lexer = NULL;
	jayfor->parser = NULL;

	// start actual useful shit here
	jayfor->scanner = scannerCreate();

	// assume second argument is a file name
	// this is a placeholder for now
	scannerReadFile(jayfor->scanner, argv[1]);

	// pass the scanned file to the lexer to tokenize
	jayfor->lexer = lexerCreate(jayfor->scanner->contents);

	return jayfor;
}

void jayforStart(Jayfor *jayfor) {
	while (jayfor->lexer->running) {
		lexerGetNextToken(jayfor->lexer);
	}

	printf("\nstart:\n");
	int i;
	for (i = 0; i < jayfor->lexer->tokenStream->size; i++) {
		Token *tok = vectorGetItem(jayfor->lexer->tokenStream, i);
		printf("%s ", tok->content);
	}
	printf("\nfinished \n");
	exit(1);

	// initialise parser after we tokenize
	jayfor->parser = parserCreate(jayfor->lexer->tokenStream);

	parserStartParsing(jayfor->parser);
}

void jayforDestroy(Jayfor *jayfor) {
	scannerDestroy(jayfor->scanner);
	lexerDestroy(jayfor->lexer);
	parserDestroy(jayfor->parser);
	free(jayfor);
}
