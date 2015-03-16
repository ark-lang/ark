#include "alloyc.h"

bool DEBUG_MODE = false;
char *OUTPUT_EXECUTABLE_NAME = "a";

static void parse_argument(CommandLineArgument *arg) {
	char argument = arg->argument[0];

	switch (argument) {
		case 'v':
			printf("alloyc version: %s\n", ALLOYC_VERSION);
			return;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'h':
			printf("Alloy-Lang Argument List\n");
			printf("\t-h,\t shows a help menu\n");
			printf("\t-v,\t shows current version\n");
			printf("\t-d,\t logs extra debug information\n");
			printf("\n");
			return;
		case 'o':
			if (!arg->nextArgument) {
				errorMessage("error: missing filename after '-o'");
			}
			OUTPUT_EXECUTABLE_NAME = arg->nextArgument;
			break;
		default:
			errorMessage("error: unrecognized command line option '-%s'\n", arg->argument);
			break;
	}
}

AlloyCompiler *createAlloyCompiler(int argc, char** argv) {
	AlloyCompiler *self = safeMalloc(sizeof(*self));
	self->filename = NULL;
	self->scanner = NULL;
	self->lexer = NULL;
	self->parser = NULL;
	self->compiler = NULL;
	self->semantic = NULL;

	// not enough args just throw an error
	if (argc <= 1) {
		errorMessage("no input files");
		return self;
	}

	int i;
	// i = 1 to ignore first arg
	for (i = 1; i < argc; i++) {
		if (argv[i][0] == '-') {
			CommandLineArgument arg;

			// remove the -
			size_t len = strlen(argv[i]) - 1;
			char temp[len];
			memcpy(temp, &argv[i][len], 1);
			temp[len] = '\0';

			// set argument stuff
			arg.argument = temp;
			arg.nextArgument = NULL;
			if (argv[i + 1] != NULL) {
				arg.nextArgument = argv[i + 1];
			}

			// multiple arguments needed for -o, consume twice
			// todo make this cleaner for when we expand
			if (!strcmp(arg.argument, "o") || !strcmp(arg.argument, "e")) {
				i += 2;
			}

			// parse the argument
			parse_argument(&arg);
		}
		else if (strstr(argv[i], ".ay")) {
			self->filename = argv[i];
		}
		else {
			errorMessage("argument not recognized: %s\n", argv[i]);
		}
	}

	return self;
}

void startAlloyCompiler(AlloyCompiler *self) {
	// filename is null, so we should exit
	// out of here
	if (self->filename == NULL) {
		return;
	}

	// start actual useful shit here
	self->scanner = createScanner();
	scanFile(self->scanner, self->filename);

	// lex file
	self->lexer = createLexer(self->scanner->contents);
	while (self->lexer->running) {
		getNextToken(self->lexer);
	}

	// calculate time for lexing
	self->lexer->timer = clock() - self->lexer->timer;
	self->lexer->secondsTaken = ((double) self->lexer->timer) / CLOCKS_PER_SEC;
	self->lexer->msTaken = self->lexer->secondsTaken * 1000;

	// initialise parser after we tokenize
	self->parser = createParser(self->lexer->tokenStream);
	parseTokenStream(self->parser);
	
	// calculate time for parsing
	self->parser->timer = clock() - self->parser->timer;
	self->parser->secondsTaken = ((double) self->parser->timer) / CLOCKS_PER_SEC;
	self->parser->msTaken = self->parser->secondsTaken * 1000;

	// failed parsing stage
	if (self->parser->exitOnError) {
		return; // don't do stuff after this
	}

	// semantic analysis, not started yet
	self->semantic = createSemanticAnalyser(self->parser->parseTree);
	startSemanticAnalysis(self->semantic);

	// compilation stage
	self->compiler = createCompiler();
	startCompiler(self->compiler, self->parser->parseTree);

	// calculate time of compiler
	self->compiler->timer = clock() - self->compiler->timer;
	self->compiler->secondsTaken = ((double) self->compiler->timer) / CLOCKS_PER_SEC;
	self->compiler->msTaken = self->compiler->secondsTaken * 1000;

	primaryMessage("Lexing: %.6f, Parsing: %.6f, Compilation: %.6f", self->lexer->secondsTaken, self->parser->secondsTaken, self->compiler->secondsTaken);

}

void destroyAlloyCompiler(AlloyCompiler *self) {
	destroyScanner(self->scanner);
	destroyLexer(self->lexer);
	destroyParser(self->parser);
	destroySemanticAnalyser(self->semantic);
	destroyCompiler(self->compiler);
	free(self);
}
