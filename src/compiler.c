#include "compiler.h"

bool DEBUG_MODE = false;	// default is no debug
bool OUTPUT_C = false;		// default is no c output
bool VERBOSE_MODE = false;
#ifdef ENABLE_LLVM
	bool LLVM_CODEGEN = false;
#endif
char *OUTPUT_EXECUTABLE_NAME = "main"; // default is main
char *COMPILER = "cc";
char *ADDITIONAL_COMPILER_ARGS = "-g -Wall -std=c99 -fno-builtin -Wno-incompatible-pointer-types-discards-qualifiers";

void help() {
	printf("Usage: alloyc [options] files...\n");
	printf("Options:\n");
	printf("  -h                  Shows this help menu\n");
	printf("  -v                  Verbose compilation\n");
	printf("  -d                  Logs extra debug information\n");
	printf("  -o <file>           Place the output into <file>\n");
	printf("  -c <file>           Will keep the output C code\n");
	printf("  --compiler <name>   Sets the C compiler to <name> (default: CC)\n");
	printf("  --version           Shows current version\n");
#ifdef ENABLE_LLVM
	printf("  --llvm              Sets the codegen backend to LLVM (default: C)\n");
#endif
}

void version() {
	printf("%s %s\n", COMPILER_NAME, COMPILER_VERSION);
#ifdef ENABLE_LLVM
	printf("LLVM backend: yes\n");
#else
	printf("LLVM backend: no\n");
#endif
}

static void parseArgument(CommandLineArgument *arg) {
	if (!strcmp(arg->argument, VERSION_ARG)) {
		version();
		exit(0);
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
		exit(0);
	}
	else if (!strcmp(arg->argument, COMPILER_ARG)) {
		if (!arg->nextArgument) {
			errorMessage("Missing compiler command after '" COMPILER_ARG "'");
		}
		COMPILER = arg->nextArgument;
	}
	else if (!strcmp(arg->argument, OUTPUT_ARG)) {
		if (!arg->nextArgument) {
			errorMessage("Missing filename after '" OUTPUT_ARG "'");
		}
		OUTPUT_EXECUTABLE_NAME = arg->nextArgument;
	}
#ifdef ENABLE_LLVM
	else if (!strcmp(arg->argument, LLVM_ARG)) {
		LLVM_CODEGEN = true;
	}
#endif
	else {
		errorMessage("Unrecognized command line option '%s'", arg->argument);
	}
}

Compiler *createCompiler(int argc, char** argv) {
	// not enough arguments just throw an error
	if (argc <= 1) {
		errorMessage("No input files");
		return NULL;
	}

	Compiler *self = safeMalloc(sizeof(*self));
	self->lexer = NULL;
	self->parser = NULL;
	self->generator = NULL;
#ifdef ENABLE_LLVM
	self->generatorLLVM = NULL;
#endif
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

			parseArgument(&arg);
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
			errorMessage("Argument not recognized: %s", argv[i]);
		}
	}

	return self;
}

void startCompiler(Compiler *self) {
	if (!self->sourceFiles || self->sourceFiles->size == 0) {
		return;
	}

	// lex file
	verboseModeMessage("Started lexing");
	self->lexer = createLexer(self->sourceFiles);
	startLexingFiles(self->lexer, self->sourceFiles);
	if (self->lexer->failed) {
		return;
	}
	verboseModeMessage("Finished lexing");

	// initialise parser after we tokenize
	verboseModeMessage("Started parsing");
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
#ifdef ENABLE_LLVM
	if (LLVM_CODEGEN) {
		self->generatorLLVM = createLLVMCodeGenerator(self->sourceFiles);
		startLLVMCodeGeneration(self->generatorLLVM);
	}
	else {
#endif
		self->generator = createCCodeGenerator(self->sourceFiles);
		startCCodeGeneration(self->generator);
#ifdef ENABLE_LLVM
	}
#endif
}

void destroyCompiler(Compiler *self) {
	if (self) {
		if (self->lexer) destroyLexer(self->lexer);
		if (self->parser) destroyParser(self->parser);
		if (self->generator) destroyCCodeGenerator(self->generator);
#ifdef ENABLE_LLVM
		if (self->generatorLLVM) destroyLLVMCodeGenerator(self->generatorLLVM);
#endif
		if (self->semantic) destroySemanticAnalyzer(self->semantic);
		destroyVector(self->sourceFiles);
		free(self);
		verboseModeMessage("Destroyed compiler");
	}
}
