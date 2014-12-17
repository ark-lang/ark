#include "jayfor.h"

Jayfor *createJayfor(int argc, char** argv) {
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
	jayfor->scanner = createScanner();

	// assume last argument is a file name
	// this is a placeholder for now
	char *filename = argv[argc-1];
	// check if the filename doesn't end with j4
	if(strstr(filename, ".j4") != NULL) {
		scanFile(jayfor->scanner, filename);
	} else {
		printf(KRED "error: not a valid j4 file.\n" KNRM);
		exit(1);
	}

	// pass the scanned file to the lexer to tokenize
	jayfor->lexer = createLexer(jayfor->scanner->contents);
	jayfor->compiler = NULL;
	jayfor->j4vm = NULL;

	return jayfor;
}

void startJayfor(Jayfor *jayfor) {
	while (jayfor->lexer->running) {
		getNextToken(jayfor->lexer);
	}

	// initialise parser after we tokenize
	jayfor->parser = createParser(jayfor->lexer->tokenStream);

	startParsingTokenStream(jayfor->parser);

	jayfor->compiler = createCompiler();
	startCompiler(jayfor->compiler, jayfor->parser->parseTree);

	jayfor->j4vm = createJayforVM();
	startJayforVM(jayfor->j4vm, jayfor->compiler->bytecode, jayfor->compiler->globalCount + 1);
}

void destroyJayfor(Jayfor *jayfor) {
	destroyScanner(jayfor->scanner);
	destroyLexer(jayfor->lexer);
	destroyParser(jayfor->parser);
	destroyCompiler(jayfor->compiler);
	free(jayfor);
}
