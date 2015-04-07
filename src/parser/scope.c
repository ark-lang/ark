#include "scope.h"

Scope *createScope() {
	Scope *scope = malloc(sizeof(*scope));
	scope->pointers = createVector(VECTOR_LINEAR);
	return scope;
}

void destroyScope(Scope *scope) {
	free(scope);
}
