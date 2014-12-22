#include "jayfor.h"

bool DEBUG_MODE = false;

static void parseArgument(char *arg) {
	char argument = arg[0];
	switch (argument) {
		case 'v':
			printf(KGRN "Jayfor Version: %s\n" KNRM, JAYFOR_VERSION);
			exit(1);
			break;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'h':
			printf(KYEL "JAYFOR Argument List\n");
			printf("\t-h,\t shows a help menu\n");
			printf("\t-v,\t shows current version\n");
			printf("\t-d,\t logs extra debug information\n");
			printf("\n" KNRM);
			exit(1);
			break;
		default:
			printf(KRED "error: unrecognized argument %c\n" KNRM, argument);
			exit(1);
			break;
	}
}

Jayfor *createJayfor(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	}

	char *filename = NULL;
	Jayfor *jayfor = malloc(sizeof(*jayfor));

	if (!jayfor) {
		perror("malloc: failed to allocate memory for JAYFOR");
		exit(1);
	}

	int i;
	// i = 1 to ignore first arg
	for (i = 1; i < argc; i++) {
		if (argv[i][0] == '-') {
			size_t len = strlen(argv[i]) - 1;
			char temp[len];
			memcpy(temp, &argv[i][len], 1);
			temp[len] = '\0';
			parseArgument(temp);
		}
		else if (strstr(argv[i], ".j4")) {
			filename = argv[i];
		}
		else {
			printf(KRED "error: argument not recognized: %s\n" KNRM, argv[i]);
		}
	}

	// just in case.
	jayfor->scanner = NULL;
	jayfor->lexer = NULL;
	jayfor->parser = NULL;

	// start actual useful shit here
	jayfor->scanner = createScanner();
	scanFile(jayfor->scanner, filename);

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
