#include "alloyc.h"

bool DEBUG_MODE = false;	// default is no debug
char *OUTPUT_EXECUTABLE_NAME = "main"; // default is main
char *COMPILER = "gcc"; // default is GCC
bool OUTPUT_C = false;	// default is no c output

void help() {
	printf("Alloy-Lang Argument List\n");
	printf("  -help\t\t\tShows this help menu\n");
	printf("  -v\t\t\tShows current version\n");
	printf("  -d\t\t\tLogs extra debug information\n");
	printf("  -c\t\t\tKeep the generated C code\n");
	printf("  -compiler <cc>\tCompiles generated code with <cc>\n");
	printf("  -o <file>\t\tPlace the output into <file>\n");
	printf("\n");
}

static void parse_argument(CommandLineArgument *arg) {
	char *argument = arg->argument;

	if (!strcmp(argument, VERSION_ARG)) {
		printf("Alloy Compiler Version: %s\n", ALLOYC_VERSION);
	}
	else if (!strcmp(argument, DEBUG_MODE_ARG)) {
		DEBUG_MODE = true;
	}
	else if (!strcmp(argument, OUTPUT_C_ARG)) {
		OUTPUT_C = true;
	}
	else if (!strcmp(argument, HELP_ARG)) {
		help();
		return;
	}
	else if (!strcmp(argument, COMPILER_ARG)) {
		COMPILER = arg->nextArgument;
	}
	else if (!strcmp(argument, OUTPUT_ARG)) {
		if (!arg->nextArgument) {
			errorMessage("missing filename after '-o'");
		}
		OUTPUT_EXECUTABLE_NAME = arg->nextArgument;
	}
	else {
		errorMessage("unrecognized command line option '%s'\n", arg->argument);
	}
}

AlloyCompiler *createAlloyCompiler(int argc, char** argv) {
	AlloyCompiler *self = safeMalloc(sizeof(*self));
	self->lexer = NULL;
	self->parser = NULL;
	self->compiler = NULL;
	self->sourceFiles = createVector();

	// not enough arguments just throw an error
	if (argc <= 1) {
		errorMessage("no input files");
		return NULL;
	}

	int i;
	// i = 1 to ignore first arg
	for (i = 1; i < argc; i++) {
		if (argv[i][0] == '-') {
			CommandLineArgument arg;

			// set argument stuff
			arg.argument = argv[i];
			arg.nextArgument = NULL;
			if (argv[i + 1] != NULL) {
				arg.nextArgument = argv[i + 1];
			}

			// make this cleaner, I see a shit ton of if statements
			// in our future.
			if (!strcmp(arg.argument, OUTPUT_ARG) ||
				!strcmp(arg.argument, COMPILER_ARG)) {
				i++;
			}

			// parse the argument
			parse_argument(&arg);
		}
		else if (strstr(argv[i], ".ay")) {
			pushBackItem(self->sourceFiles, createSourceFile(argv[i]));
		}
		else {
			errorMessage("argument not recognized: %s\n", argv[i]);
		}
	}

	return self;
}

void startAlloyCompiler(AlloyCompiler *self) {
	if (!self->sourceFiles || self->sourceFiles->size == 0) {
		return;
	}

	// lex file
	self->lexer = createLexer(self->sourceFiles);
	startLexingFiles(self->lexer, self->sourceFiles);

	// initialise parser after we tokenize
	self->parser = createParser();
	startParsingSourceFiles(self->parser, self->sourceFiles);
	
	// failed parsing stage
	if (self->parser->exitOnError) {
		return; // don't do stuff after this
	}

	// compilation stage
	self->compiler = createCompiler(self->sourceFiles);
	startCompiler(self->compiler);
}

void destroyAlloyCompiler(AlloyCompiler *self) {
	if (self) {
		if (self->lexer) destroyLexer(self->lexer);
		if (self->parser) destroyParser(self->parser);
		if (self->compiler) destroyCompiler(self->compiler);
		destroyVector(self->sourceFiles);
		free(self);
	}
}
