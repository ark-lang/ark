#include "jayfor.h"

Jayfor *jayforCreate(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf("error: no input files\n");
		exit(1);
	}

	Jayfor *jayfor = malloc(sizeof(*jayfor));
	if (!jayfor) {
		perror("malloc: failed to allocate memory for JAYFOR");
		exit(1);
	}
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
