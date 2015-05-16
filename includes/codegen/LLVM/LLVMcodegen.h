#ifndef __LLVM_CODEGEN_H
#define __LLVM_CODEGEN_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/BitWriter.h>

#include "util.h"
#include "parser.h"
#include "vector.h"
#include "hashmap.h"

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
	 * Our index in the abstract syntax
	 * tree
	 */
	int currentNode;
} LLVMCodeGenerator;

/**
 * Creates an instance of the code generator
 * @param  sourceFiles the source files to codegen for
 * @return             the instance we created
 */
LLVMCodeGenerator *createLLVMCodeGenerator(Vector *sourceFiles);

/**
 * This is pretty much where the magic happens, this will
 * start the code gen
 * @param self the code gen instance
 */
void startLLVMCodeGeneration(LLVMCodeGenerator *self);

/**
 * Destroys the given code gen instance
 * @param self the code gen instance
 */
void destroyLLVMCodeGenerator(LLVMCodeGenerator *self);

#endif
