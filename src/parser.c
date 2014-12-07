#include "parser.h"

Parser *parserCreate(Vector *tokenStream) {
	Parser *parser = malloc(sizeof(*parser));
	if (!parser) {
		perror("malloc: failed to allocate memory for parser");
		exit(1);
	}
	parser->tokenStream = tokenStream;
	parser->tokenIndex = 0;
	parser->parsing = true;
	return parser;
}

void parserStartParsing(Parser *parser) {
	while (parser->parsing) {
		Token *tok = vectorGetItem(parser->tokenStream, parser->tokenIndex);
		if (tok->type == END_OF_FILE) {
			parser->parsing = false;
			break;
		}
		printf("%s\n", tok->content);
		parser->tokenIndex += 1;
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