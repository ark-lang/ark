#ifndef __CODE_GEN_H
#define __CODE_GEN_H

/**
 * The code generator! This works in 2 stages, the first stage it
 * will try and emit code for all of the macros we're given. The second
 * stage is where we generate the "meat" of the program, all of the statements,
 * structures, declarations, etc are generated.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "parser.h"
#include "vector.h"
#include "hashmap.h"

/**
 * Constants that are generated in the
 * code should be defined here so we can
 * change them in the future if we need to
 * etc...
 */
#define SPACE_CHAR " "
#define OPEN_BRACKET "("
#define CLOSE_BRACKET ")"
#define OPEN_BRACE "{"
#define CLOSE_BRACE "}"
#define CONST_KEYWORD "const"
#define ASTERISKS "*"
#ifdef WINDOWS
	#define NEWLINE "\r\n"
#else
	#define NEWLINE "\n"
#endif
#define TAB "\t"
#define EQUAL_SYM "="
#define SEMICOLON ";"
#define COMMA_SYM ","

/**
 * This defined whether the code we 
 * produce is somewhat readable, or
 * all minified
 */
#define COMPACT_CODE_GEN false

#if COMPACT_CODE_GEN == false
	#define CC_NEWLINE NEWLINE
#else
	#define CC_NEWLINE " "
#endif

/**
 * This is the format of generated impl functions.
 * First one is the struct name, second one the function name.
 */
#define GENERATED_IMPL_FUNCTION_FORMAT "__%s_%s"

/**
 * The two types of state that
 * we can be in, this makes it cleaner
 * than switching between files or writing
 * one after the other.
 */
typedef enum {
	WRITE_HEADER_STATE,
	WRITE_SOURCE_STATE
} WriteState;

typedef struct {
	/**
	 * The current abstract syntax tree being
	 * generated for
	 */
	Vector *abstractSyntaxTree;
	
	/**
	 * All of the source files we're generating code
	 * for
	 */
	Vector *sourceFiles;

	/**
	 * The current source file to
	 * generate code for
	 */
	SourceFile *currentSourceFile;
	
	/**
	 * The current write state, i.e
	 * are we writing to the header
	 * or source file
	 */
	WriteState writeState;

	sds linkerFlags;

	/**
	 * Our index in the abstract syntax
	 * tree
	 */
	int currentNode;
} CCodeGenerator;

/**
 * Creates an instance of the code generator
 * @param  sourceFiles the source files to codegen for
 * @return             the instance we created
 */
CCodeGenerator *createCCodeGenerator(Vector *sourceFiles);

/**
 * This is pretty much where the magic happens, this will
 * start the code gen
 * @param self the code gen instance
 */
void startCCodeGeneration(CCodeGenerator *self);

/** MACROS */

/**
 * Generates the code for all the macros, the code generator
 * is currently in 2 passes, the first will generate the code
 * for all of the macros, the second will generate code for other
 * statements
 * @param self the code gen instance
 */
void generateMacrosC(CCodeGenerator *self);

/**
 * Destroys the given code gen instance
 * @param self the code gen instance
 */
void destroyCCodeGenerator(CCodeGenerator *self);

#endif // __CODE_GEN_H
