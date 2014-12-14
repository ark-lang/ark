#ifndef CODEGEN_H
#define CODEGEN_H

#include <stdio.h>
#include <stdlib.h>

typedef struct {
	
} Backend;

Backend *backendCreate();

void backendDestroy(Backend *backend);

#endif // CODEGEN_H