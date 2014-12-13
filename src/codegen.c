#include "codegen.h"

Backend *backendCreate() {
	Backend *self = malloc(sizeof(*self));
	return self;
}

void backendDestroy(Backend *backend) {
	free(backend);
}