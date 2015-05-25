#ifndef __VARIABLE_H
#define __VARIABLE_H

#include <llvm-c/Core.h>

#include "util.h"
#include "hashmap.h"

typedef struct {
    char *name;
    LLVMValueRef value;
} VariableReference;

VariableReference *createVariableRef(char *name);

void destroyVariableRef(VariableReference *ref);

#endif // __VARIABLE_H