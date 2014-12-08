#include "parser.h"

static char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

Parser *parserCreate(Vector *tokenStream) {
	Parser *parser = malloc(sizeof(*parser));
	if (!parser) {
		perror("malloc: failed to allocate memory for parser");
		exit(1);
	}
	parser->tokenStream = tokenStream;
	parser->parseTree = vectorCreate();
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

Token *parserExpectType(Parser *parser, TokenType type) {
	Token *tok = parserPeekAhead(parser, 1);
	if (tok->type == type) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

// simple testing parse thing
void parserParseInteger(Parser *parser) {
	Token *integerName = parserExpectType(parser, IDENTIFIER);
	Token *nextToken = parserPeekAhead(parser, 1);
	VariableDefineNode vdn;

	// todo finish this
	if (nextToken->type == OPERATOR && nextToken->content[0] == ';') {
		// not assigning a value
		vdn.type = INTEGER;
		vdn.name = integerName;
	}
	else if (nextToken->type == OPERATOR && nextToken->content[0] == '=') {
		// assigning a value
	}
	else {

	}
}

void parserStartParsing(Parser *parser) {
	while (parser->parsing) {
		// get current token
		Token *tok = vectorGetItem(parser->tokenStream, parser->tokenIndex);

		switch (tok->type) {
			case IDENTIFIER:
				if (!strcmp(tok->content, "int")) {
					parserParseInteger(parser);
				}
				break;
			case END_OF_FILE:
				parser->parsing = false;
				break;
		}

		// print contents
		printf("%s \t\t\t %s\n", tok->content, getTokenName(tok));
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