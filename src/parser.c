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
	ExpressionNode *expr = malloc(sizeof(*expr));
	expr->value = NULL;
	expr->lhand = NULL;
	expr->rhand = NULL;

	if (!expr) {
		perror("malloc: failed to allocate memory for ExpressionNode");
		exit(1);
	}
	return expr;
}

VariableDefineNode *createVariableDefineNode() {
	VariableDefineNode *vdn = malloc(sizeof(*vdn));
	vdn->name = NULL;

	if (!vdn) {
		perror("malloc: failed to allocate memory for VariableDefineNode");
		exit(1);
	}
	return vdn;
}

VariableDeclareNode *createVariableDeclareNode() {
	VariableDeclareNode *vdn = malloc(sizeof(*vdn));
	vdn->vdn = NULL;
	vdn->expr = NULL;

	if (!vdn) {
		perror("malloc: failed to allocate memory for VariableDeclareNode");
		exit(1);
	}
	return vdn;
}

FunctionArgumentNode *createFunctionArgumentNode() {
	FunctionArgumentNode *fan = malloc(sizeof(*fan));
	fan->name = NULL;
	fan->value = NULL;

	if (!fan) {
		perror("malloc: failed to allocate memory for FunctionArgumentNode");
		exit(1);
	}
	return fan;
}

FunctionCalleeNode *createFunctionCalleeNode() {
	FunctionCalleeNode *fcn = malloc(sizeof(*fcn));
	fcn->callee = NULL;
	fcn->args = NULL;

	if (!fcn) {
		perror("malloc: failed to allocate memory for FunctionCalleeNode");
		exit(1);
	}
	return fcn;
}

BlockNode *createBlockNode() {
	BlockNode *bn = malloc(sizeof(*bn));
	bn->statements = NULL;

	if (!bn) {
		perror("malloc: failed to allocate memory for BlockNode");
		exit(1);
	}
	return bn;
}

FunctionPrototypeNode *createFunctionPrototypeNode() {
	FunctionPrototypeNode *fpn = malloc(sizeof(*fpn));
	fpn->args = NULL;
	fpn->name = NULL;

	if (!fpn) {
		perror("malloc: failed to allocate memory for FunctionPrototypeNode");
		exit(1);
	}
	return fpn;
}

FunctionNode *createFunctionNode() {
	FunctionNode *fn = malloc(sizeof(*fn));
	fn->fpn = NULL;
	fn->body = NULL;

	if (!fn) {
		perror("malloc: failed to allocate memory for FunctionNode");
		exit(1);
	}
	return fn;
}

void destroyExpressionNode(ExpressionNode *expr) {
	if (expr->lhand != NULL) {
		destroyExpressionNode(expr->lhand);
	}
	if (expr->rhand != NULL) {
		destroyExpressionNode(expr->rhand);
	}
	free(expr);
}

void destroyVariableDefineNode(VariableDefineNode *vdn) {
	free(vdn);
}

void destroyVariableDeclareNode(VariableDeclareNode *vdn) {
	if (vdn->vdn != NULL) {
		destroyVariableDefineNode(vdn->vdn);
	}
	if (vdn->expr != NULL) {
		destroyExpressionNode(vdn->expr);
	}
	free(vdn);
}

void destroyFunctionArgumentNode(FunctionArgumentNode *fan) {
	if (fan->value != NULL) {
		destroyExpressionNode(fan->value);
	}
	free(fan);
}

void destroyBlockNode(BlockNode *bn) {
	if (bn->statements != NULL) {
		vectorDestroy(bn->statements);
	}
	free(bn);
}

void destroyFunctionPrototypeNode(FunctionPrototypeNode *fpn) {
	if (fpn->args != NULL) {
		vectorDestroy(fpn->args);
	}
	free(fpn);
}

void destroyFunctionNode(FunctionNode *fn) {
	if (fn->fpn != NULL) {
		destroyFunctionPrototypeNode(fn->fpn);
	}
	if (fn->body != NULL) {
		destroyBlockNode(fn->body);
	}
	free(fn);
}

void destroyFunctionCalleeNode(FunctionCalleeNode *fcn) {
	if (fcn->args != NULL) {
		vectorDestroy(fcn->args);
	}
	free(fcn);
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
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *parserExpectContent(Parser *parser, char *content) {
	Token *tok = parserPeekAhead(parser, 1);
	if (!strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, tok->content, content);
		exit(1);
	}
}

Token *parserExpectTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = parserPeekAhead(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *parserMatchType(Parser *parser, TokenType type) {
	Token *tok = parserPeekAhead(parser, 0);
	if (tok->type == type) {
		return parserConsumeToken(parser);
	}
	else {
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *parserMatchContent(Parser *parser, char *content) {
	Token *tok = parserPeekAhead(parser, 0);
	if (!strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, tok->content, content);
		exit(1);
	}
}

Token *parserMatchTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = parserPeekAhead(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return parserConsumeToken(parser);
	}
	else {
		printf("%d:%d expected %s but found `%s`\n", tok->pos->lineNumber, tok->pos->charNumber, TOKEN_NAMES[type], tok->content);
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

	printf(KRED "error: failed to parse expression, STRING & NUMBER only supported\n" KNRM);
	exit(1);
}

void printCurrentToken(Parser *parser) {
	Token *tok = parserPeekAhead(parser, 0);
	printf(KYEL "current token is type: %s, value: %s\n" KNRM, TOKEN_NAMES[tok->type], tok->content);
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
		prepareNode(parser, dec, VARIABLE_DEC_NODE);

		// match a semi colon
		parserMatchTypeAndContent(parser, SEPARATOR, ";");
	}
	else if (parserTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		// consume the semi colon
		parserConsumeToken(parser);

		// create variable define node
		VariableDefineNode *def = createVariableDefineNode();
		def->type = dataTypeRaw;
		def->name = variableNameToken;
		prepareNode(parser, def, VARIABLE_DEF_NODE);
	}
	else {
		// error message
		printf(KRED "error: missing a semi colon or assignment\n" KNRM);
		exit(1);
	}
}

BlockNode *parserParseBlock(Parser *parser) {
	BlockNode *block = createBlockNode();

	parserMatchTypeAndContent(parser, SEPARATOR, "{");
	
	do {
		// check if block is empty before we try parse some statements
		if (parserTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
			parserConsumeToken(parser);
			break;
		}

		parserParseStatements(parser);
	}
	while (true);

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
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				parserConsumeToken(parser);
				break;
			}

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
					printf(KRED "error: trailing comma at the end of argument list\n" KNRM);
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
		prepareNode(parser, fn, FUNCTION_NODE);
	}
	else {
		printf(KRED "error: no parameter list provided\n" KNRM);
		exit(1);
	}
}

void parserParseFunctionCall(Parser *parser) {
	// consume function name
	Token *callee = parserMatchType(parser, IDENTIFIER);

	if (parserTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		parserConsumeToken(parser);	// eat open bracket

		Vector *args = vectorCreate();

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				parserConsumeToken(parser);
				break;
			}

			ExpressionNode *expr = parserParseExpression(parser);

			FunctionArgumentNode *arg = createFunctionArgumentNode();
			arg->value = expr;

			if (parserTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (parserTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					printf(KRED "error: trailing comma at the end of argument list\n" KNRM);
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

		// consume semi colon
		parserMatchTypeAndContent(parser, SEPARATOR, ";");

		// woo we got the function
		FunctionCalleeNode *fcn = createFunctionCalleeNode();
		fcn->callee = callee;
		fcn->args = args;
		prepareNode(parser, fcn, FUNCTION_CALLEE_NODE);
	}
}

void parserParseReturnStatement(Parser *parser) {

}

void parserParseStatements(Parser *parser) {
	// ret keyword	
	if (parserTokenTypeAndContent(parser, IDENTIFIER, "ret", 0)) {
		printf("swag\n");
	}
	else if (parserTokenType(parser, IDENTIFIER, 0)) {
		if (parserTokenTypeAndContent(parser, SEPARATOR, "(", 1)) {
			Token *tok = parserPeekAhead(parser, 0);
			printf("parsing function call to %s\n", tok->content);
			parserParseFunctionCall(parser);
		}
		else {
			printf("no function what is this\n");
		}
	}
	else {
		Token *tok = parserPeekAhead(parser, 0);
		printf("error: unrecognized token %s(%s)\n", tok->content, TOKEN_NAMES[tok->type]);
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
					printf(KRED "error: unrecognized identifier found: `%s`\n" KNRM, tok->content);
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
	printf(KRED "error: invalid data type specified: %s!\n" KNRM, tok->content);
	exit(1);
}

void prepareNode(Parser *parser, void *data, NodeType type) {
	Node *node = malloc(sizeof(*node));
	node->data = data;
	node->type = type;
	vectorPushBack(parser->parseTree, node);
}

void removeNode(Node *node) {
	/**
	 * This could probably be a lot more cleaner
	 */
	if (node->data != NULL) {
		switch (node->type) {
			case EXPRESSION_NODE:
				destroyExpressionNode(node->data);
				break;
			case VARIABLE_DEF_NODE: 
				destroyVariableDefineNode(node->data);
				break;
			case VARIABLE_DEC_NODE:
				destroyVariableDeclareNode(node->data);
				break;
			case FUNCTION_ARG_NODE:
				destroyFunctionArgumentNode(node->data);
				break;
			case FUNCTION_NODE:
				destroyFunctionNode(node->data);
				break;
			case FUNCTION_PROT_NODE:
				destroyFunctionPrototypeNode(node->data);
				break;
			case BLOCK_NODE:
				destroyBlockNode(node->data);
				break;
			case FUNCTION_CALLEE_NODE:
				destroyFunctionCalleeNode(node->data);
				break;
			default:
				printf(KYEL "attempting to remove unrecognized node(%d)?\n" KNRM, node->type);
				break;
		}
	}
	free(node);
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

	for (i = 0; i < parser->parseTree->size; i++) {
		Node *node = vectorGetItem(parser->parseTree, i);
		removeNode(node);
	}
	vectorDestroy(parser->parseTree);

	// finally destroy parser
	free(parser);
}