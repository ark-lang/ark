#include "parser.h"

/** List of token names */
static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void"
};

/** UTILITY FOR NODES */

ExpressionNode *createExpressionNode() {
	return malloc(sizeof(ExpressionNode));
}

VariableDefineNode *createVariableDefineNode() {
	return malloc(sizeof(VariableDefineNode));
}

VariableDeclareNode *createVariableDeclareNode() {
	return malloc(sizeof(VariableDeclareNode));
}

FunctionArgumentNode *createFunctionArgumentNode() {
	return malloc(sizeof(FunctionArgumentNode));
}

FunctionNode *createFunctionNode() {
	return malloc(sizeof(FunctionNode));
}

BlockNode *createBlockNode() {
	return malloc(sizeof(BlockNode));
}

FunctionPrototypeNode *createFunctionPrototypeNode() {
	return malloc(sizeof(FunctionPrototypeNode));
}

/** END NODE FUNCTIONS */

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

ExpressionNode *parserParseExpression(Parser *parser) {
	ExpressionNode *expr = createExpressionNode(); // the final expression

	// number literal
	if (parserTokenType(parser, NUMBER, 0)) {
		expr->type = 'N';
		expr->value = parserConsumeToken(parser);
		return expr;
	}

	// string literal
	if (parserTokenType(parser, STRING, 0)) {
		expr->type = 'S';
		expr->value = parserConsumeToken(parser);
		return expr;
	}

	// todo: variable reference
	// int x = y;

	// actual expressions e.g (5 + 5) - (10 ^ 3)

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

		// create variable define node
		VariableDefineNode *def = createVariableDefineNode();
		def->type = dataTypeRaw;
		def->name = variableNameToken;

		// parses the expression we're assigning to
		ExpressionNode *expr = parserParseExpression(parser);

		// create the variable declare node
		VariableDeclareNode *dec = createVariableDeclareNode();
		dec->vdn = def;
		dec->expr = expr;

		// match a semi colon
		parserMatchTypeAndContent(parser, SEPARATOR, ";");

		// print out for debug
		printf("just parsed an %s declaration!!\nLet's see what we missed:\n", variableDataType->content);
	}
	else if (parserTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		// consume the semi colon
		parserConsumeToken(parser);

		// create variable define node
		VariableDefineNode *def = createVariableDefineNode();
		def->type = dataTypeRaw;
		def->name = variableNameToken;

		// print out for debug
		printf("hey we've just parsed an %s definition!!!\nLet's see what's left...\n", variableDataType->content);
	}
	else {
		// error message
		printf("missing a semi colon or assignment, found\n");
		printCurrentToken(parser);
		exit(1);
	}
}

BlockNode *parserParseBlock(Parser *parser) {
	BlockNode *block = createBlockNode();

	parserMatchTypeAndContent(parser, SEPARATOR, "{");
	
	do {
		parserConsumeToken(parser);
	}
	while (parserTokenTypeAndContent(parser, SEPARATOR, "}", 0));

	return block;
}

void parserParseFunctionPrototype(Parser *parser) {
	parserMatchType(parser, IDENTIFIER);	// consume the fn keyword

	Token *functionName = parserMatchType(parser, IDENTIFIER);
	Vector *args = vectorCreate(); // null for now till I add arg parsing

	FunctionPrototypeNode *fpn = createFunctionPrototypeNode();
	fpn->args = args;
	fpn->name = functionName;

	// parameter list
	if (parserTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		parserConsumeToken(parser);

		do {
			Token *argDataType = parserMatchType(parser, IDENTIFIER);
			DataType argRawDataType = parserTokenTypeToDataType(parser, argDataType);
			Token *argName = parserMatchType(parser, IDENTIFIER);

			FunctionArgumentNode *arg = createFunctionArgumentNode();
			arg->type = argRawDataType;
			arg->name = argName;
			arg->value = NULL;

			if (parserTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
				parserConsumeToken(parser);

				// default expression
				ExpressionNode *expr = parserParseExpression(parser);
				arg->value = expr;
				vectorPushBack(args, arg);

				if (parserTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					parserConsumeToken(parser);
				}
				else if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					parserConsumeToken(parser); // eat closing parenthesis
					break;
				}
			}
			else if (parserTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					printf("OMG A TRAILING COMMA\n");
					exit(1);
				}
				parserConsumeToken(parser); // eat the comma
				vectorPushBack(args, arg);
			}
			else if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				parserConsumeToken(parser); // eat closing parenthesis
				vectorPushBack(args, arg);
				break;
			}
		}
		while (true);

		// consume colon 
		parserMatchTypeAndContent(parser, OPERATOR, ":");

		// consume return type
		Token *functionReturnType = parserMatchType(parser, IDENTIFIER);
		
		// convert to raw data type, also good for errors
		DataType returnTypeRaw = parserTokenTypeToDataType(parser, functionReturnType);

		// store the return type in the function prototype node
		fpn->ret = returnTypeRaw;

		// start block
		BlockNode *body = parserParseBlock(parser);

		// create the function because we have a body declared
		FunctionNode *fn = createFunctionNode();
		fn->fpn = fpn;
		fn->body = body;

		printf("we've finished parsing a function that returns %s\n", functionReturnType->content);
	}
	else {
		printf("WHERES THE PARAMETER LIST LEBOWSKI?\n");
	}

	int i;
	for (i = 0; i < fpn->args->size; i++) {
		FunctionArgumentNode *arg = vectorGetItem(fpn->args, i);
		char *content = arg->value != NULL ? arg->value->value->content : "NULL";
		printf("%s %s = %s\n", DATA_TYPES[arg->type], arg->name->content, content);
		if (arg->value != NULL) {
			free(arg->value);
		}
		free(arg);
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
				if (!strcmp(tok->content, "fn")) {
					parserParseFunctionPrototype(parser);
				} 
				else if (parserIsTokenDataType(parser, tok)) {
					parserParseVariable(parser);
				}
				else {
					printf("Unrecognized identifier found: `%s`\n", tok->content);
					exit(1);
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