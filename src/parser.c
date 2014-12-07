#include "parser.h"

Parser *parserCreate(Vector *tokenStream) {
	Parser *parser = malloc(sizeof(*parser));
	parser->tokenStream = tokenStream;
	return parser;
}

NumberExpr exprParseNumber() {
	NumberExpr result = {
		.value
	}
}

void parserDestroy(Parser *parser) {
	int i;
	for (i = 0; i < parser->tokenStream->size; i++) {
		Token *tok = vectorGetItem(parser->tokenStream, i);
		free(tok->content);
		tokenDestroy(tok);
	}
	vectorDestroy(parser->tokenStream);
	free(parser);
}