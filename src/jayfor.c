#include "jayfor.h"

Jayfor *jayforCreate(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	} 

	// used as a counter for getopt()
	int argCounter;

	/* 
	 * arguments specified after each tag. e.g. -o a.out
	 * here a.out will be stored in nextArg
	 */
	char *nextArg; 
	while ((argCounter = getopt(argc, argv, "v:o:")) != -1) {
		switch(argCounter) {
			case 'v':
				printf(KYEL "v argument found.\n\n" KNRM);
				nextArg = optarg;
				printf(KYEL "argument specified with v: %s\n\n" KNRM, nextArg);
				break;
			case 'o':
				printf(KYEL "o argument found.\n\n" KNRM);
				nextArg = optarg;
				printf("argument specified with o: %s\n\n", nextArg);
				break;
			default:
				printf(KRED "invalid argument.\n\n" KNRM);
				abort();
		}
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

	// assume last argument is a file name
	// this is a placeholder for now
	char *filename = argv[argc-1];
	// check if the filename doesn't end with j4
	if(strstr(filename, ".j4") != NULL) {
		scannerReadFile(jayfor->scanner, filename);
	} else {
		printf(KRED "error: not a valid j4 file.\n" KNRM);
		exit(1);
	}

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

	jayfor->compiler = compilerCreate();
	compilerStart(jayfor->compiler);
}

void jayforDestroy(Jayfor *jayfor) {
	scannerDestroy(jayfor->scanner);
	lexerDestroy(jayfor->lexer);
	parserDestroy(jayfor->parser);
	compilerDestroy(jayfor->compiler);
	free(jayfor);
}
