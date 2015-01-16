#include "preprocessor.h"

/**
 * for now this will be a really badly implemented
 * preprocessor for simplicity-sake
 */

preprocessor *create_preprocessor(vector *token_stream) {
	preprocessor *self = safe_malloc(sizeof(*self));
	self->token_stream = token_stream;
	return self;
}

void start_preprocessing(preprocessor *self) {
	
}

void destroy_preprocessor(preprocessor *self) {
	free(self);
}
