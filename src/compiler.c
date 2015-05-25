#include "compiler.h"

bool DEBUG_MODE = false;	// default is no debug
bool VERBOSE_MODE = false;
char *OUTPUT_EXECUTABLE_NAME = "main"; // default is main
bool IGNORE_MAIN = false;
char *COMPILER = "cc";

Compiler *createCompiler(int argc, char** argv) {
	// not enough arguments just throw an error
	if (argc <= 1) {
		errorMessage("No input files");
		return NULL;
	}

	Compiler *self = safeMalloc(sizeof(*self));
	self->lexer = NULL;
	self->parser = NULL;
	self->generatorLLVM = NULL;
	self->semantic = NULL;

	char *ccEnv = getenv("CC");
	if (ccEnv != NULL && strcmp(ccEnv, ""))
		COMPILER = ccEnv;

	self->sourceFiles = setup_arguments(argc, argv);
	if (self->sourceFiles->size == 0) {
		destroyCompiler(self);
		return NULL;
	}

	return self;
}

void startCompiler(Compiler *self) {
	if (!self->sourceFiles || self->sourceFiles->size == 0) {
		printf("not running \n");
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

	self->generatorLLVM = createLLVMCodeGenerator(self->sourceFiles);
	startLLVMCodeGeneration(self->generatorLLVM);
}

void destroyCompiler(Compiler *self) {
	if (self) {
		if (self->lexer) destroyLexer(self->lexer);
		if (self->parser) destroyParser(self->parser);
		if (self->generatorLLVM) destroyLLVMCodeGenerator(self->generatorLLVM);
		if (self->semantic) destroySemanticAnalyzer(self->semantic);
		if (self->sourceFiles) destroyVector(self->sourceFiles);
		free(self);
		verboseModeMessage("Destroyed compiler");
	}
}
