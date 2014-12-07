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

Token *parserConsumeToken(Parser *parser) {
	// return the token we are consuming, then increment token index
	return vectorGetItem(parser->tokenStream, parser->tokenIndex++);
}

Token *parserPeekAhead(Parser *parser, int ahead) {
	return vectorGetItem(parser->tokenStream, parser->tokenIndex + ahead);
}

void parserStartParsing(Parser *parser) {
	while (parser->parsing) {
		// get current token
		Token *tok = vectorGetItem(parser->tokenStream, parser->tokenIndex);
		
		// break out if we're EOF
		parser->parsing = tok->type != END_OF_FILE;

		// print contents
		printf("%s \t\t\t %s\n", tok->content, getTokenName(tok));

		// consume token
		parserConsumeToken(parser);	
	}
}

void parserDestroy(Parser *parser) {
	int i;
	for (i = 0; i < parser->tokenStream->size; i++) {
		Token *tok = vectorGetItem(parser->tokenStream, i);
		// gotta free the content that was allocated
		free(tok->content);

		// then we destroy token
		tokenDestroy(tok);
	}
	// destroy the token stream once we're done with it
	vectorDestroy(parser->tokenStream);

	// finally destroy parser
	free(parser);
}