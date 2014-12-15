#ifndef COMPILER_H
#define COMPILER_H

#include <stdio.h>
#include <stdlib.h>

#include <llvm-c/Core.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Transforms/Scalar.h>

#include "util.h"
#include "hashmap.h"

typedef struct {
	char *error;
	LLVMModuleRef module;
	LLVMBuilderRef builder;
	LLVMExecutionEngineRef engine;
	LLVMPassManagerRef passManager;
} Compiler;

Compiler *compilerCreate();

void compilerStart(Compiler *compiler);

void compilerDestroy(Compiler *compiler);

#endif // COMPILER_H