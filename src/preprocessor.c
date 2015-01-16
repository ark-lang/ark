#include "preprocessor.h"

preprocessor *create_preprocessor(char *file_content) {
	preprocessor *self = safe_malloc(sizeof(*self));
	self->file_content = file_content;
	return self;
}

char *start_preprocessing(preprocessor *self) {
	int i;
	size_t len = strlen(self->file_content);
	
	for (i = 0; i < len; i++) {
		
	}

	return NULL;
}

void destroy_preprocessor(preprocessor *self) {
	free(self);
}
