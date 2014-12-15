#include "compiler.h"

Compiler *compilerCreate() {
	Compiler *compiler = malloc(sizeof(*compiler));
	if (!compiler) {
		perror(KRED "error: failed to allocate memory for Compiler" KNRM);
		exit(1);
	}
	compiler->error = NULL;
	return compiler;
}

void compilerStart(Compiler *compiler) {
	compiler->module = LLVMModuleCreateWithName("j4");
	compiler->builder = LLVMCreateBuilder();
	compiler->engine;

	LLVMInitializeNativeTarget();
	LLVMLinkInJIT();

	compiler->error = NULL;
	if (LLVMCreateExecutionEngineForModule(&compiler->engine, compiler->module, &compiler->error)) {
		fprintf(stderr, "%s\n", compiler->error);
		LLVMDisposeMessage(compiler->error);
		exit(1);
	}

	compiler->passManager = LLVMCreateFunctionPassManagerForModule(compiler->module);
	LLVMAddTargetData(LLVMGetExecutionEngineTargetData(compiler->engine), compiler->passManager);
	LLVMAddPromoteMemoryToRegisterPass(compiler->passManager);
	LLVMAddInstructionCombiningPass(compiler->passManager);
	LLVMAddReassociatePass(compiler->passManager);
	LLVMAddGVNPass(compiler->passManager);
	LLVMAddCFGSimplificationPass(compiler->passManager);
	LLVMInitializeFunctionPassManager(compiler->passManager);
}

void compilerDestroy(Compiler *compiler) {
	if (compiler != NULL) {
	    LLVMDumpModule(compiler->module);
		LLVMDisposePassManager(compiler->passManager);
	    LLVMDisposeBuilder(compiler->builder);
	    LLVMDisposeModule(compiler->module);

		free(compiler);
		compiler = NULL;
	}
}
