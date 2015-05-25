#include "LLVM/variable.h"

VariableReference *createVariableRef(char *name) {
    VariableReference *ref = safeMalloc(sizeof(*ref));
    ref->name = name;
    verboseModeMessage("Storing var reference as %s\n", ref->name);
    return ref;
}

void destroyVariableRef(VariableReference *ref) {
    free(ref);
}