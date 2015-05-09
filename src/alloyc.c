#include "alloyc.h"

bool DEBUG_MODE = false;	// default is no debug
bool OUTPUT_C = false;		// default is no c output
bool VERBOSE_MODE = false;
char *OUTPUT_EXECUTABLE_NAME = "main"; // default is main
char *COMPILER = "cc";
char *ADDITIONAL_COMPILER_ARGS = "-g -Wall -std=c99 -fno-builtin";

void help() {
	printf("Alloy-Lang Argument List\n");
	printf("  -h\t\t\tShows this help menu\n");
	printf("  -ver\t\t\tShows current version\n");
	printf("  -v\t\t\tVerbose compilation\n");
	printf("  -d\t\t\tLogs extra debug information\n");
	printf("  -o <file>\t\tPlace the output into <file>\n");
	printf("  -c <file>\t\tWill keep the output C code\n");
	printf("\n");
}

static void parse_argument(CommandLineArgument *arg) {
	if (!strcmp(arg->argument, VERSION_ARG)) {
		printf("Alloy Compiler Version: %s\n", ALLOYC_VERSION);
	}
	else if (!strcmp(arg->argument, DEBUG_MODE_ARG)) {
		DEBUG_MODE = true;
	}
	else if (!strcmp(arg->argument, OUTPUT_C_ARG)) {
		OUTPUT_C = true;
	}
	else if (!strcmp(arg->argument, VERBOSE_ARG)) {
		VERBOSE_MODE = true;
	}
	else if (!strcmp(arg->argument, HELP_ARG)) {
		help();
		return;
	}
	else if (!strcmp(arg->argument, COMPILER_ARG)) {
		COMPILER = arg->nextArgument;
	}
	else if (!strcmp(arg->argument, OUTPUT_ARG)) {
		if (!arg->nextArgument) {
			errorMessage("missing filename after '" OUTPUT_ARG "'");
		}
		OUTPUT_EXECUTABLE_NAME = arg->nextArgument;
	}
	else {
		errorMessage("unrecognized command line option '%s'\n", arg->argument);
	}
}

AlloyCompiler *createAlloyCompiler(int argc, char** argv) {
	// not enough arguments just throw an error
	if (argc <= 1) {
		errorMessage("no input files");
		return NULL;
	}

	AlloyCompiler *self = safeMalloc(sizeof(*self));
	self->lexer = NULL;
	self->parser = NULL;
	self->generator = NULL;
	self->sourceFiles = createVector(VECTOR_LINEAR);
	
	char *ccEnv = getenv("CC");
	if (ccEnv != NULL && strcmp(ccEnv, ""))
		COMPILER = ccEnv;

	// i = 1, ignores first argument
	for (int i = 1; i < argc; i++) {
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
			if (!strcmp(arg.argument, OUTPUT_ARG)
				|| !strcmp(arg.argument, COMPILER_ARG)) {
				i++; // skips the argument
			}

			// parse the argument
			parse_argument(&arg);
		}
		else if (strstr(argv[i], ".aly")) {
			SourceFile *file = createSourceFile(sdsnew(argv[i]));
			if (!file) {
				verboseModeMessage("Error when attempting to create a source file");
				return NULL;
			}
			pushBackItem(self->sourceFiles, file);
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
	if (self->lexer->failed) {
		return;
	}
	verboseModeMessage("Finished Lexing");

	// initialise parser after we tokenize
	self->parser = createParser();
	startParsingSourceFiles(self->parser, self->sourceFiles);
	verboseModeMessage("Finished parsing");
	
	// failed parsing stage
	if (self->parser->failed) {
		return; // don't do stuff after this
	}

	self->semantic = createSemanticAnalyzer(self->sourceFiles);
	startSemanticAnalysis(self->semantic);
	if (self->semantic->failed) {
		return;
	}

	// compilation stage
	self->generator = createCodeGenerator(self->sourceFiles);
	startCodeGeneration(self->generator);
}

void destroyAlloyCompiler(AlloyCompiler *self) {
	if (self) {
		if (self->lexer) destroyLexer(self->lexer);
		if (self->parser) destroyParser(self->parser);
		if (self->generator) destroyCodeGenerator(self->generator);
		destroyVector(self->sourceFiles);
		free(self);
		verboseModeMessage("Destroyed Alloy Compiler");
	}
}
