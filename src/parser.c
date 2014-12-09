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

Token *parserExpectTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = parserPeekAhead(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *parserMatchType(Parser *parser, TokenType type) {
	Token *tok = parserPeekAhead(parser, 0);
	if (tok->type == type) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *parserMatchTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = parserPeekAhead(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

bool parserTokenType(Parser *parser, TokenType type) {
	Token *tok = parserPeekAhead(parser, 1);
	return tok->type == type;
}

bool parserTokenContent(Parser *parser, char* content) {
	Token *tok = parserPeekAhead(parser, 1);
	return !strcmp(tok->content, content);
}

Expression parserParseExpression(Parser *parser) {
	Expression expr;

	if (parserTokenType(parser, NUMBER)) {
		parserMatchTypeAndContent(parser, OPERATOR, "=");

		expr.type = 'N';
		expr.value = parserConsumeToken(parser);
		printf("parsed an expression\n");

		parserMatchTypeAndContent(parser, SEPARATOR, ";");

		return expr;
	}

	printf("failed to parse expression!\n");
	exit(1);
}

void printCurrentToken(Parser *parser) {
	Token *tok = parserPeekAhead(parser, 0);
	printf("current token is type: %s, value: %s\n", TOKEN_NAMES[tok->type], tok->content);
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
		vectorPushBack(parser->parseTree, &vdn);
		parserConsumeToken(parser);

		printf("theres another shebang\n");
		return;
	}
	else if (nextToken->type == OPERATOR && nextToken->content[0] == '=') {
		// assigning a value
		printf("assigning a value\n");

		// store vdn
		vdn.type = INTEGER;
		vdn.name = integerName;

		// consume the equals		
		parserConsumeToken(parser);

		// probably an expression		
		Expression expr = parserParseExpression(parser);

		// woah make it mon
		VariableDeclareNode declareNode;
		declareNode.vdn = vdn;
		declareNode.expr = &expr;
		vectorPushBack(parser->parseTree, &declareNode);
		
		printf("fuckin sorted\n");
		
		printCurrentToken(parser);
		return;
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