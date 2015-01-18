#include "preprocessor.h"

/**
 * for now this will be a really badly implemented
 * preprocessor for simplicity-sake
 *
 * I'm hoping to implement something similar to Rusts macro
 * system. Maybe macros will be prefix with an exclamation,
 * i.e
 *
 * !use file
 * !define whatever "whatever"
 *
 * However, we would actually work with the AST as opposed to
 * c's preproccessor just does a replace on all instances.
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
