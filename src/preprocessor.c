#include "preprocessor.h"

/**
 * for now this will be a really badly implemented
 * preprocessor for simplicity-sake
 */

preprocessor *create_preprocessor(vector *token_stream) {
	preprocessor *self = safe_malloc(sizeof(*self));
	self->token_stream = token_stream;
	self->token_index = 0;
	return self;
}

void parse_macro(vector *macro) {
	// todo
}

void start_preprocessing(preprocessor *self) {
	// todo
}

void destroy_preprocessor(preprocessor *self) {
	free(self);
}
