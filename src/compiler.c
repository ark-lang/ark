#include "compiler.h"

Compiler *createCompiler() {
	Compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	return self;
}

void startCompiler(Compiler *self, Vector *ast) {
	self->ast = ast;
}

void destroyCompiler(Compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
