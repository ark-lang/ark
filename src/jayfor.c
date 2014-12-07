#include "jayfor.h"

Jayfor *jayforCreate(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf("error: no input files\n");
		exit(1);
	}

	Jayfor *jayfor = malloc(sizeof(*jayfor));
	jayfor->scanner = scannerCreate();
	scannerReadFile(jayfor->scanner, argv[1]);
	jayfor->lexer = lexerCreate(jayfor->scanner->contents);

	return jayfor;
}

void jayforStart(Jayfor *jayfor) {
	do {
		lexerGetNextToken(jayfor->lexer);
	}
	while (jayfor->lexer->running);
}

void jayforDestroy(Jayfor *jayfor) {
	scannerDestroy(jayfor->scanner);
	lexerDestroy(jayfor->lexer);
	free(jayfor);
}
