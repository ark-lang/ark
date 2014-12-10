#include "parser.h"

static char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

static char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void"
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

Token *parserExpectContent(Parser *parser, char *content) {
	Token *tok = parserPeekAhead(parser, 1);
	if (!strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", content, tok->content);
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

Token *parserMatchContent(Parser *parser, char *content) {
	Token *tok = parserPeekAhead(parser, 0);
	if (!strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", content, tok->content);
		exit(1);
	}
}

Token *parserMatchTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = parserPeekAhead(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", content, tok->content);
		exit(1);
	}
}

bool parserTokenType(Parser *parser, TokenType type, int ahead) {
	Token *tok = parserPeekAhead(parser, ahead);
	return tok->type == type;
}

bool parserTokenContent(Parser *parser, char* content, int ahead) {
	Token *tok = parserPeekAhead(parser, ahead);
	return !strcmp(tok->content, content);
}

bool parserTokenTypeAndContent(Parser *parser, TokenType type, char* content, int ahead) {
	return parserTokenType(parser, type, ahead) && parserTokenContent(parser, content, ahead);
}

Expression parserParseExpression(Parser *parser) {
	Expression expr; // the final expression
	if (parserTokenType(parser, NUMBER, 0)) {
		expr.type = 'N';
		expr.value = parserConsumeToken(parser);
		printf("parsed an expression\n");
		return expr;
	}
	if (parserTokenType(parser, STRING, 0)) {
		expr.type = 'S';
		expr.value = parserConsumeToken(parser);
		printf("parsed an expression\n");
		return expr;
	}
	printf("failed to parse expression, we found this:");
	printCurrentToken(parser);
	exit(1);
}

void printCurrentToken(Parser *parser) {
	Token *tok = parserPeekAhead(parser, 0);
	printf("current token is type: %s, value: %s\n", TOKEN_NAMES[tok->type], tok->content);
}

void parserParseVariable(Parser *parser) {
	// TYPE NAME = 5;
	// TYPE NAME;

	// consume the int data type
	Token *variableDataType = parserMatchType(parser, IDENTIFIER);

	// convert the data type for enum
	DataType dataTypeRaw = parserTokenTypeToDataType(parser, variableDataType);

	// name of the variable
	Token *variableNameToken = parserMatchType(parser, IDENTIFIER);

	if (parserTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
		// consume the equals sign
		parserConsumeToken(parser);

		printCurrentToken(parser);

		// create variable define node
		VariableDefineNode def;
		def.type = dataTypeRaw;
		def.name = variableNameToken;

		// parses the expression we're assigning to
		Expression expr = parserParseExpression(parser);

		// create the variable declare node
		VariableDeclareNode dec;
		dec.vdn = def;
		dec.expr = &expr;

		// match a semi colon
		parserMatchTypeAndContent(parser, SEPARATOR, ";");

		// print out for debug
		printf("just parsed an %s declaration!!\nLet's see what we missed:\n", variableDataType->content);
		printCurrentToken(parser);
	}
	else if (parserTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		// consume the semi colon
		parserConsumeToken(parser);

		// create variable define node
		VariableDefineNode def;
		def.type = dataTypeRaw;
		def.name = variableNameToken;

		// print out for debug
		printf("hey we've just parsed an %s definition!!!\nLet's see what's left...\n", variableDataType->content);
		printCurrentToken(parser);
	}
	else {
		// error message
		printf("missing a semi colon or assignment, found\n");
		printCurrentToken(parser);
		exit(1);
	}
}

Block parserParseBlock(Parser *parser) {
	Block block;

	parserMatchTypeAndContent(parser, SEPARATOR, "{");
	
	do {
		printf("statement\n");
		parserConsumeToken(parser);
	}
	while (parserTokenTypeAndContent(parser, SEPARATOR, "}", 0));

	printCurrentToken(parser);
	
	return block;
}

void parserParseFunctionPrototype(Parser *parser) {
	parserMatchType(parser, IDENTIFIER);	// consume the fn keyword

	Token *functionName = parserMatchType(parser, IDENTIFIER);
	Vector *args = NULL;

	FunctionPrototypeNode fpn;
	fpn.args = args;
	fpn.name = functionName;

	// parameter list
	if (parserTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		do {
			//TODO: implement arguments,
			// this would be variable defines
			// except no semicolon.
			printf("list item\n");
			parserConsumeToken(parser);
		}
		while (parserTokenTypeAndContent(parser, SEPARATOR, ")", 0));
		printf("end of arg list\n");

		// consume colon 
		parserMatchTypeAndContent(parser, OPERATOR, ":");

		// consume return type
		Token *functionReturnType = parserMatchType(parser, IDENTIFIER);
		
		// convert to raw data type, also good for errors
		DataType returnTypeRaw = parserTokenTypeToDataType(parser, functionReturnType);

		// store the return type in the function prototype node
		fpn.ret = returnTypeRaw;

		// start block
		Block body = parserParseBlock(parser);

		// create the function because we have a body declared
		FunctionNode fn;
		fn.fpn = fpn;
		fn.body = &body;

		printf("we've finished parsing a function that returns %s\n", functionReturnType->content);
	}
	else {
		printf("WHERES THE PARAMETER LIST LEBOWSKI?\n");
	}
}

void parserStartParsing(Parser *parser) {
	while (parser->parsing) {
		// get current token
		Token *tok = vectorGetItem(parser->tokenStream, parser->tokenIndex);

		switch (tok->type) {
			case IDENTIFIER:
				// parse a variable if we have a variable
				// given to us
				if (parserIsTokenDataType(parser, tok)) {
					parserParseVariable(parser);
				}
				else if (!strcmp(tok->content, "fn")) {
					parserParseFunctionPrototype(parser);
				}
				break;
			case END_OF_FILE:
				parser->parsing = false;
				break;
		}
	}
}

bool parserIsTokenDataType(Parser *parser, Token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return true;
		}
	}
	return false;
}

DataType parserTokenTypeToDataType(Parser *parser, Token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return i;
		}
	}
	printf("Invalid data type %s!\n", tok->content);
	exit(1);
}

void parserDestroy(Parser *parser) {
	int i;
	for (i = 0; i < parser->tokenStream->size; i++) {
		Token *tok = vectorGetItem(parser->tokenStream, i);
			
		// then we destroy token
		tokenDestroy(tok);
	}

	// destroy the token stream once we're done with it
	vectorDestroy(parser->tokenStream);
	vectorDestroy(parser->parseTree);

	// finally destroy parser
	free(parser);
}