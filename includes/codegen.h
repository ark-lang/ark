#ifndef CODEGEN_H
#define CODEGEN_H

#include <llvm-c/Core.h>

typedef struct {
	LLVMModuleRef module;
	LLVMBuilderRef builder;
} Backend;

Backend *backendCreate();

void backendDestroy(Backend *backend);

#endif // CODEGEN_H