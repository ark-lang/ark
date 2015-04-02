#include "compiler.h"

void generateFunctionCode(Compiler *self, FunctionDecl *decl) {
	const int numOfParams = decl->signature->parameters->paramList->size;

	LLVMTypeRef params[numOfParams];

	int i;
	for (i = 0; i < numOfParams; i++) {
		ParameterSection *param = getVectorItem(decl->signature->parameters->paramList, i);
		LLVMTypeRef type = getTypeRef(param->type);
		if (type) {
			params[i] = type;
		}
	}

	LLVMTypeRef returnType = getTypeRef(decl->signature->type);
	if (returnType) {
		LLVMTypeRef funcRet = LLVMFunctionType(returnType, params, numOfParams, false);
		LLVMValueRef func = LLVMAddFunction(self->module, decl->signature->name, funcRet);

		LLVMBasicBlockRef funcBlock = LLVMAppendBasicBlock(func, decl->signature->name);
		LLVMPositionBuilderAtEnd(self->builder, funcBlock);
	}

}