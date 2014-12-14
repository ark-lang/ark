#include "codegen.h"

Backend *backendCreate() {
	Backend *self = malloc(sizeof(*self));
	if (!self) {
		perror("malloc: failed to allocate memory for backend");
		exit(1);
	}
	return self;
}

void backendDestroy(Backend *backend) {
	free(backend);
}