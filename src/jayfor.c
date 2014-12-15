#include "jayfor.h"

Jayfor *jayforCreate(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	} 

	// used as a counter for getopt()
	int c;
	/* arguments specified after each tag. e.g. -o a.out
	 * here a.out will be stored in opt_arg
	 */
	char *opt_arg; 
	while((c = getopt (argc, argv, "v:o:")) != -1 ) {
		switch(c) {
			case 'v':
				printf("v argument found.\n\n");
				opt_arg = optarg;
				printf("argument specified with v: %s\n\n", opt_arg);
				break;
			case 'o':
				printf("o argument found.\n\n");
				opt_arg = optarg;
				printf("argument specified with o: %s\n\n", opt_arg);
				break;
			default:
				printf("invalid argument.\n\n");
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
}

void jayforDestroy(Jayfor *jayfor) {
	scannerDestroy(jayfor->scanner);
	lexerDestroy(jayfor->lexer);
	parserDestroy(jayfor->parser);
	free(jayfor);
}
