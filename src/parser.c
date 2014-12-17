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
	"void", "char", "tup"
};

// static const char* NODE_NAMES[] = {
// 	"expression_node", "variable_def_node",
// 	"variable_dec_node", "function_arg_node",
// 	"function_node", "function_prot_node",
// 	"block_node", "function_callee_node",
// 	"function_ret_node", "for_loop_node",
// 	"variable_reassign_node"
// };

/** UTILITY FOR NODES */

VariableReassignNode *createVariableReassignNode() {
	VariableReassignNode *vrn = malloc(sizeof(*vrn));
	if (!vrn) {
		perror("malloc: failed to allocate memory for Variable Reassign Node");
		exit(1);
	}
	vrn->name = NULL;
	vrn->expr = NULL;
	return vrn;
}

StatementNode *createStatementNode() {
	StatementNode *sn = malloc(sizeof(*sn));
	if (!sn) {
		perror("malloc: failed to allocate memory for Statement Node");
		exit(1);
	}
	sn->data = NULL;
	sn->type = 0;
	return sn;
}

FunctionReturnNode *createFunctionReturnNode() {
	FunctionReturnNode *frn = malloc(sizeof(*frn));
	if (!frn) {
		perror("malloc: failed to allocate memory for Function Return Node");
		exit(1);
	}
	frn->returnVals = NULL;
	return frn;
}

ExpressionNode *createExpressionNode() {
	ExpressionNode *expr = malloc(sizeof(*expr));
	if (!expr) {
		perror("malloc: failed to allocate memory for ExpressionNode");
		exit(1);
	}
	expr->value = NULL;
	expr->lhand = NULL;
	expr->rhand = NULL;
	return expr;
}

BooleanExpressionNode *createBooleanExpressionNode() {
    BooleanExpressionNode *boolExpr = malloc(sizeof(*boolExpr));
    if (!boolExpr) {
        perror("malloc: failed to allocate memory for BooleanExpressionNode");
        exit(1);
    }
    boolExpr->expr = NULL;
    boolExpr->lhand = NULL;
    boolExpr->rhand = NULL;
    return boolExpr;
}

VariableDefineNode *createVariableDefineNode() {
	VariableDefineNode *vdn = malloc(sizeof(*vdn));
	if (!vdn) {
		perror("malloc: failed to allocate memory for VariableDefineNode");
		exit(1);
	}
	vdn->name = NULL;
	return vdn;
}

VariableDeclareNode *createVariableDeclareNode() {
	VariableDeclareNode *vdn = malloc(sizeof(*vdn));
	if (!vdn) {
		perror("malloc: failed to allocate memory for VariableDeclareNode");
		exit(1);
	}
	vdn->vdn = NULL;
	vdn->expr = NULL;
	return vdn;
}

FunctionArgumentNode *createFunctionArgumentNode() {
	FunctionArgumentNode *fan = malloc(sizeof(*fan));
	if (!fan) {
		perror("malloc: failed to allocate memory for FunctionArgumentNode");
		exit(1);
	}
	fan->name = NULL;
	fan->value = NULL;
	return fan;
}

FunctionCalleeNode *createFunctionCalleeNode() {
	FunctionCalleeNode *fcn = malloc(sizeof(*fcn));
	if (!fcn) {
		perror("malloc: failed to allocate memory for FunctionCalleeNode");
		exit(1);
	}
	fcn->callee = NULL;
	fcn->args = NULL;
	return fcn;
}

BlockNode *createBlockNode() {
	BlockNode *bn = malloc(sizeof(*bn));
	if (!bn) {
		perror("malloc: failed to allocate memory for BlockNode");
		exit(1);
	}
	bn->statements = NULL;
	return bn;
}

FunctionPrototypeNode *createFunctionPrototypeNode() {
	FunctionPrototypeNode *fpn = malloc(sizeof(*fpn));
	if (!fpn) {
		perror("malloc: failed to allocate memory for FunctionPrototypeNode");
		exit(1);
	}
	fpn->args = NULL;
	fpn->name = NULL;
	return fpn;
}

FunctionNode *createFunctionNode() {
	FunctionNode *fn = malloc(sizeof(*fn));
	if (!fn) {
		perror("malloc: failed to allocate memory for FunctionNode");
		exit(1);
	}
	fn->fpn = NULL;
	fn->body = NULL;
	return fn;
}

ForLoopNode *createForLoopNode() {
	ForLoopNode *fln = malloc(sizeof(*fln));
	if (!fln) {
		perror("malloc: failed to allocate memory for ForLoopNode");
		exit(1);
	}
	return fln;
}

void destroyVariableReassignNode(VariableReassignNode *vrn) {
	if (vrn != NULL) {
		if (vrn->expr != NULL) {
			destroyExpressionNode(vrn->expr);
		}
		free(vrn);
	}
}

void destroyForLoopNode(ForLoopNode *fln) {
	if (fln != NULL) {
		free(fln);
		fln = NULL;
	}
}

void destroyStatementNode(StatementNode *sn) {
	if (sn != NULL) {
		if (sn->data != NULL) {
			switch (sn->type) {
				case VARIABLE_DEF_NODE:
					destroyVariableDefineNode(sn->data);
					break;
				case VARIABLE_DEC_NODE:
					destroyVariableDeclareNode(sn->data);
					break;
				case FUNCTION_CALLEE_NODE:
					destroyFunctionCalleeNode(sn->data);
					break;
				case FUNCTION_RET_NODE:
					destroyFunctionNode(sn->data);
					break;
				case VARIABLE_REASSIGN_NODE:
					destroyVariableReassignNode(sn->data);
					break;
				default: break;
			}
		}
		free(sn);
		sn = NULL;
	}
}

void destroyFunctionReturnNode(FunctionReturnNode *frn) {
	if (frn != NULL) {
		if (frn->returnVals != NULL) {
			int i;
			for (i = 0; i < frn->returnVals->size; i++) {
				ExpressionNode *temp = getItemFromVector(frn->returnVals, i);
				if (temp != NULL) {
					destroyExpressionNode(temp);
				}
			}
			destroyVector(frn->returnVals);
		}
		free(frn);
		frn = NULL;
	}
}

void destroyExpressionNode(ExpressionNode *expr) {
	if (expr != NULL) {
		if (expr->lhand != NULL) {
			destroyExpressionNode(expr->lhand);
		}
		if (expr->rhand != NULL) {
			destroyExpressionNode(expr->rhand);
		}
		free(expr);
		expr = NULL;
	}
}

void destroyVariableDefineNode(VariableDefineNode *vdn) {
	if (vdn != NULL) {
		free(vdn);
		vdn = NULL;
	}
}

void destroyVariableDeclareNode(VariableDeclareNode *vdn) {
	if (vdn != NULL) {
		if (vdn->vdn != NULL) {
			destroyVariableDefineNode(vdn->vdn);
		}
		if (vdn->expr != NULL) {
			destroyExpressionNode(vdn->expr);
		}
		free(vdn);
		vdn = NULL;
	}
}

void destroyFunctionArgumentNode(FunctionArgumentNode *fan) {
	if (fan != NULL) {
		if (fan->value != NULL) {
			destroyExpressionNode(fan->value);
		}
		free(fan);
		fan = NULL;
	}
}

void destroyBlockNode(BlockNode *bn) {
	if (bn != NULL) {
		if (bn->statements != NULL) {
			destroyVector(bn->statements);
		}
		free(bn);
		bn = NULL;
	}
}

void destroyFunctionPrototypeNode(FunctionPrototypeNode *fpn) {
	if (fpn != NULL) {
		if (fpn->args != NULL) {
			int i;
			for (i = 0; i < fpn->args->size; i++) {
				StatementNode *sn = getItemFromVector(fpn->args, i);
				if (sn != NULL) {
					destroyStatementNode(sn);
				}
			}
			destroyVector(fpn->args);
		}
		free(fpn);
		fpn = NULL;
	}
}

void destroyFunctionNode(FunctionNode *fn) {
	if (fn != NULL) {
		if (fn->fpn != NULL) {
			destroyFunctionPrototypeNode(fn->fpn);
		}
		if (fn->body != NULL) {
			destroyBlockNode(fn->body);
		}
		if (fn->ret != NULL) {
			destroyVector(fn->ret);
		}
		free(fn);
		fn = NULL;
	}
}

void destroyFunctionCalleeNode(FunctionCalleeNode *fcn) {
	if (fcn != NULL) {
		if (fcn->args != NULL) {
			destroyVector(fcn->args);
		}
		free(fcn);
		fcn = NULL;
	}
}

void destroyBooleanExpressionNode(BooleanExpressionNode *ben) {
    if (ben != NULL) {
        if(ben->lhand != NULL) {
            destroyBooleanExpressionNode(ben->lhand);
        }
        if(ben->rhand != NULL) {
            destroyBooleanExpressionNode(ben->rhand);
        }
        free(ben);
        ben = NULL;
    }
}

/** END NODE FUNCTIONS */

Parser *createParser(Vector *tokenStream) {
	Parser *parser = malloc(sizeof(*parser));
	if (!parser) {
		perror("malloc: failed to allocate memory for parser");
		exit(1);
	}
	parser->tokenStream = tokenStream;
	parser->parseTree = createVector();
	parser->tokenIndex = 0;
	parser->parsing = true;
	return parser;
}

Token *consumeToken(Parser *parser) {
	// return the token we are consuming, then increment token index
	return getItemFromVector(parser->tokenStream, parser->tokenIndex++);
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	return getItemFromVector(parser->tokenStream, parser->tokenIndex + ahead);
}

Token *expectTokenType(Parser *parser, TokenType type) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (tok->type == type) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *expectTokenContent(Parser *parser, char *content) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (!strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", tok->content, content);
		exit(1);
	}
}

Token *expectTokenTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *matchTokenType(Parser *parser, TokenType type) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (tok->type == type) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

Token *matchTokenContent(Parser *parser, char *content) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (!strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", tok->content, content);
		exit(1);
	}
}

Token *matchTokenTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		printf("expected %s but found `%s`\n", TOKEN_NAMES[type], tok->content);
		exit(1);
	}
}

bool checkTokenType(Parser *parser, TokenType type, int ahead) {
	Token *tok = peekAtTokenStream(parser, ahead);
	return tok->type == type;
}

bool checkTokenContent(Parser *parser, char* content, int ahead) {
	Token *tok = peekAtTokenStream(parser, ahead);
	return !strcmp(tok->content, content);
}

bool checkTokenTypeAndContent(Parser *parser, TokenType type, char* content, int ahead) {
	return checkTokenType(parser, type, ahead) && checkTokenContent(parser, content, ahead);
}

char parseOperand(Parser *parser) {
	Token *tok = peekAtTokenStream(parser, 0);
	char tokChar = tok->content[0];

	switch (tokChar) {
		case '+': consumeToken(parser); return tokChar;
		case '-': consumeToken(parser); return tokChar;
		case '*': consumeToken(parser); return tokChar;
		case '/': consumeToken(parser); return tokChar;
		case '%': consumeToken(parser); return tokChar;
		case '>': consumeToken(parser); return tokChar;
		case '<': consumeToken(parser); return tokChar;
		case '^': consumeToken(parser); return tokChar;
		default:
			printf(KRED "error: invalid operator ('%c') specified\n" KNRM, tok->content[0]);
			exit(1);
			break;
	}
}

StatementNode *parserParseForLoopNode(Parser *parser) {
	/**
	 * for int x:(0, 10, 2) {
	 * 
	 * }
	 */

	// for token
	matchTokenTypeAndContent(parser, IDENTIFIER, "for");					// FOR
	
	Token *dataType = matchTokenType(parser, IDENTIFIER);					// DATA_TYPE
	DataType dataTypeRaw = matchTokenTypeToDataType(parser, dataType);
	
	Token *indexName = matchTokenType(parser, IDENTIFIER);					// INDEX_NAME

	matchTokenTypeAndContent(parser, OPERATOR, ":");						// PARAMS

	ForLoopNode *fln = createForLoopNode();
	fln->type = dataTypeRaw;
	fln->indexName = indexName;
	fln->params = createVector();

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		int paramCount = 0;

		do {
			if (paramCount > 3) {
				printf(KRED "error: for loop has one too many arguments %d\n" KNRM, paramCount);
				exit(1);
			}
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				if (paramCount < 2) {
					printf(KRED "error: for loop expects a maximum of 3 arguments, you have %d\n" KNRM, paramCount);
					exit(1);
				}
				consumeToken(parser);
				break;
			}

			if (checkTokenType(parser, IDENTIFIER, 0)) {
				pushBackVectorItem(fln->params, consumeToken(parser));
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						printf(KRED "error: trailing comma in for loop declaration!\n" KNRM);
						exit(1);
					}
					consumeToken(parser);
				}
			}
			else if (checkTokenType(parser, NUMBER, 0)) {
				pushBackVectorItem(fln->params, consumeToken(parser));	
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						printf(KRED "error: trailing comma in for loop declaration!\n" KNRM);
						exit(1);
					}
					consumeToken(parser);
				}
			}
			// it's an expression probably
			else if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
				ExpressionNode *expr = parseExpressionNode(parser);
				pushBackVectorItem(fln->params, expr);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						printf(KRED "error: trailing comma in for loop declaration!\n" KNRM);
						exit(1);
					}
					consumeToken(parser);
				}
			}
			else {
				printf(KRED "error: expected a number or variable in for loop parameters, found:\n" KNRM);
				printCurrentToken(parser);
				exit(1);
			}

			paramCount++;
		}
		while (true);	
	
		fln->body = parseBlockNode(parser);

		StatementNode *sn = createStatementNode();
		sn->type = FOR_LOOP_NODE;
		sn->data = fln;
		return sn;
	}

	printf(KRED "failed to parse for loop\n" KNRM);
	exit(1);
}

ExpressionNode *parseExpressionNode(Parser *parser) {
	ExpressionNode *expr = createExpressionNode(); // the final expression

	// number literal
	if (checkTokenType(parser, NUMBER, 0)) {
		expr->type = EXPR_NUMBER;
		expr->value = consumeToken(parser);
		return expr;
	}
	// string literal
	if (checkTokenType(parser, STRING, 0)) {
		expr->type = EXPR_STRING;
		expr->value = consumeToken(parser);
		return expr;
	}
	// character
	if (checkTokenType(parser, CHARACTER, 0)) {
		expr->type = EXPR_CHARACTER;
		expr->value = consumeToken(parser);
		return expr;
	}
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		expr->type = EXPR_VARIABLE;
		expr->value = consumeToken(parser);
		return expr;
	}
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);
		expr->type = EXPR_PARENTHESIS;
		expr->lhand = parseExpressionNode(parser);
		expr->operand = parseOperand(parser);
		expr->rhand = parseExpressionNode(parser);
		if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
			consumeToken(parser);
			return expr;
		}
		printf(KRED "error: missing closing parenthesis on expression\n" KNRM);
		exit(1);
	}
    if(checkTokenTypeAndContent(parser, OPERATOR, "!", 0)) {
        consumeToken(parser);
        expr->type = EXPR_LOGICAL_OPERATOR;
    }

	printf(KRED "error: failed to parse expression, only character, string and numbers are supported\n" KNRM);
	printCurrentToken(parser);
	exit(1);
}

void printCurrentToken(Parser *parser) {
	Token *tok = peekAtTokenStream(parser, 0);
	printf(KYEL "current token is type: %s, value: %s\n" KNRM, TOKEN_NAMES[tok->type], tok->content);
}

void *parseVariableNode(Parser *parser, bool global) {
	// TYPE NAME = 5;
	// TYPE NAME;

	// consume the int data type
	Token *variableDataType = matchTokenType(parser, IDENTIFIER);

	// convert the data type for enum
	DataType dataTypeRaw = matchTokenTypeToDataType(parser, variableDataType);

	// name of the variable
	Token *variableNameToken = matchTokenType(parser, IDENTIFIER);

	if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
		// consume the equals sign
		consumeToken(parser);

		// create variable define node
		VariableDefineNode *def = createVariableDefineNode();
		def->type = dataTypeRaw;
		def->name = variableNameToken;

		// parses the expression we're assigning to
		ExpressionNode *expr = parseExpressionNode(parser);

		// create the variable declare node
		VariableDeclareNode *dec = createVariableDeclareNode();
		dec->vdn = def;
		dec->expr = expr;

		// match a semi colon
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}

		if (global) {
			prepareNode(parser, dec, VARIABLE_DEC_NODE);
			return dec;
		}
		StatementNode *sn = createStatementNode();
		sn->data = dec;
		sn->type = VARIABLE_DEC_NODE;
		return sn;
	}
	else {
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}

		// create variable define node
		VariableDefineNode *def = createVariableDefineNode();
		def->type = dataTypeRaw;
		def->name = variableNameToken;
		
		if (global) {
			prepareNode(parser, def, VARIABLE_DEF_NODE);
			return def;
		}
		StatementNode *sn = createStatementNode();
		sn->data = def;
		sn->type = VARIABLE_DEF_NODE;
		return sn;
	}
}

BlockNode *parseBlockNode(Parser *parser) {
	BlockNode *block = createBlockNode();
	block->statements = createVector();

	matchTokenTypeAndContent(parser, SEPARATOR, "{");
	
	do {
		// check if block is empty before we try parse some statements
		if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
			consumeToken(parser);
			break;
		}

		pushBackVectorItem(block->statements, parseStatementNode(parser));
	}
	while (true);

	// int i;
	// for (i = 0; i < block->statements->size; i++) {
	// 	StatementNode *sn = getItemFromVector(block->statements, i);
	// 	printf("%d = %s\n", i, NODE_NAMES[sn->type]);
	// }
	// printf("\n");

	return block;
}

FunctionNode *parseFunctionNode(Parser *parser) {
	matchTokenType(parser, IDENTIFIER);	// consume the fn keyword

	Token *functionName = matchTokenType(parser, IDENTIFIER); // name of function
	Vector *args = createVector(); // null for now till I add arg parsing

	// Create function signature
	FunctionPrototypeNode *fpn = createFunctionPrototypeNode();
	fpn->args = args;
	fpn->name = functionName;

	// parameter list
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			Token *argDataType = matchTokenType(parser, IDENTIFIER);
			DataType argRawDataType = matchTokenTypeToDataType(parser, argDataType);
			Token *argName = matchTokenType(parser, IDENTIFIER);

			FunctionArgumentNode *arg = createFunctionArgumentNode();
			arg->type = argRawDataType;
			arg->name = argName;
			arg->value = NULL;

			if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
				consumeToken(parser);

				// default expression
				ExpressionNode *expr = parseExpressionNode(parser);
				arg->value = expr;
				pushBackVectorItem(args, arg);

				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					consumeToken(parser);
				}
				else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser); // eat closing parenthesis
					break;
				}
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					printf(KRED "error: trailing comma at the end of argument list\n" KNRM);
					exit(1);
				}
				consumeToken(parser); // eat the comma
				pushBackVectorItem(args, arg);
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser); // eat closing parenthesis
				pushBackVectorItem(args, arg);
				break;
			}
		}
		while (true);

		FunctionNode *fn = createFunctionNode();
		fn->ret = createVector();
		fn->numOfReturnValues = 0;
		fn->isTuple = false;

		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);
		}
		else {
			printf(KRED "error: function signature missing colon\n" KNRM);
			exit(1);
		}

		// START OF TUPLE
		if (checkTokenTypeAndContent(parser, OPERATOR, "<", 0)) {
			consumeToken(parser);
			fn->isTuple = true;

			do {
				if (checkTokenTypeAndContent(parser, OPERATOR, ">", 0)) {
					if (fn->numOfReturnValues < 1) {
						printf(KRED "error: function expects a return type\n" KNRM);
						exit(1);
					}
					consumeToken(parser); // eat
					break;
				}

				if (checkTokenType(parser, IDENTIFIER, 0)) {
					Token *tok = consumeToken(parser);
					if (checkTokenTypeIsValidDataType(parser, tok)) {
						DataType rawDataType = matchTokenTypeToDataType(parser, tok);
						pushBackVectorItem(fn->ret, &rawDataType);
						fn->numOfReturnValues++;
					}
					else {
						printf(KRED "error: invalid data type specified: `%s`\n" KNRM, tok->content);
						exit(1);
					}
					if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
						if (checkTokenTypeAndContent(parser, OPERATOR, ">", 1)) {
							printf(KRED "error: trailing comma in function declaraction\n" KNRM);
							exit(1);
						}
						consumeToken(parser);
					}
				}
			}
			while (true);
		}
		else if (checkTokenType(parser, IDENTIFIER, 0)) {
			Token *returnType = consumeToken(parser);
			DataType rawDataType = matchTokenTypeToDataType(parser, returnType);
			pushBackVectorItem(fn->ret, &rawDataType);
			fn->numOfReturnValues += 1;
		}
		else {
			printf(KRED "error: function declaration return type expected, found this:\n" KNRM);
			printCurrentToken(parser);
			exit(1);
		}

		// start block
		BlockNode *body = parseBlockNode(parser);
		fn->fpn = fpn;
		fn->body = body;
		prepareNode(parser, fn, FUNCTION_NODE);

		return fn;
	}
	else {
		printf(KRED "error: no parameter list provided\n" KNRM);
		exit(1);
	}

	// just in case we fail to parse, free this shit
	free(fpn);
	fpn = NULL;
}

FunctionCalleeNode *parseFunctionNodeCall(Parser *parser) {
	// consume function name
	Token *callee = matchTokenType(parser, IDENTIFIER);

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);	// eat open bracket

		Vector *args = createVector();

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			ExpressionNode *expr = parseExpressionNode(parser);

			FunctionArgumentNode *arg = createFunctionArgumentNode();
			arg->value = expr;

			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					printf(KRED "error: trailing comma at the end of argument list\n" KNRM);
					exit(1);
				}
				consumeToken(parser); // eat the comma
				pushBackVectorItem(args, arg);
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser); // eat closing parenthesis
				pushBackVectorItem(args, arg);
				break;
			}
		}
		while (true);

		// consume semi colon
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}

		// woo we got the function
		FunctionCalleeNode *fcn = createFunctionCalleeNode();
		fcn->callee = callee;
		fcn->args = args;
		prepareNode(parser, fcn, FUNCTION_CALLEE_NODE);
		return fcn;
	}

	printf(KRED "error: failed to parse function call\n" KNRM);
	exit(1);
}

FunctionReturnNode *parserParseReturnStatement(Parser *parser) {
	// consume the return keyword
	matchTokenTypeAndContent(parser, IDENTIFIER, "ret");

	FunctionReturnNode *frn = createFunctionReturnNode();
	frn->returnVals = createVector();
	frn->numOfReturnValues = 0;

	if (checkTokenTypeAndContent(parser, OPERATOR, "<", 0)) {
		consumeToken(parser);

		do {
			if (checkTokenTypeAndContent(parser, OPERATOR, ">", 0)) {
				consumeToken(parser);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
					consumeToken(parser);
				}
				return frn;
			}

			ExpressionNode *expr = parseExpressionNode(parser);
			pushBackVectorItem(frn->returnVals, expr);
			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, OPERATOR, ">", 1)) {
					printf(KRED "error: trailing comma in return statement\n" KNRM);
					exit(1);
				}
				consumeToken(parser);
				frn->numOfReturnValues++;
			}
		}
		while (true);
	}
	else {
		// only one return type
		ExpressionNode *expr = parseExpressionNode(parser);
		pushBackVectorItem(frn->returnVals, expr);
		frn->numOfReturnValues++;

		// consume semi colon if present
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		return frn;
	}

	printf(KRED "error: failed to parse return statement\n" KNRM);
	exit(1);
}

StatementNode *parseStatementNode(Parser *parser) {
	// ret keyword	
	if (checkTokenTypeAndContent(parser, IDENTIFIER, "ret", 0)) {
		StatementNode *sn = createStatementNode();
		sn->data = parserParseReturnStatement(parser); 
		sn->type = FUNCTION_RET_NODE;
		return sn;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, "for", 0)) {
		return parserParseForLoopNode(parser);
	}
	else if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *tok = peekAtTokenStream(parser, 0);
		
		// variable reassignment
		if (checkTokenTypeAndContent(parser, OPERATOR, "=", 1)) {
			StatementNode *sn = createStatementNode();
			sn->data = parseReassignmentStatementNode(parser);
			sn->type = VARIABLE_REASSIGN_NODE;
			return sn;
		}
		// function call
		else if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 1)) {
			StatementNode *sn = createStatementNode();
			sn->data = parseFunctionNodeCall(parser);
			sn->type = FUNCTION_CALLEE_NODE;
			return sn;
		}
		// local variable
		else if (checkTokenTypeIsValidDataType(parser, tok)) {
			return parseVariableNode(parser, false);
		}
		// fuck knows
		else {
			printf("error: unrecognized identifier %s\n", tok->content);
			exit(1);
		}
	}

	Token *tok = peekAtTokenStream(parser, 0);
	printf(KRED "error: unrecognized token %s(%s)\n" KNRM, tok->content, TOKEN_NAMES[tok->type]);
	exit(1);
}

VariableReassignNode *parseReassignmentStatementNode(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *variableName = consumeToken(parser);

		if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
			consumeToken(parser);

			ExpressionNode *expr = parseExpressionNode(parser);

			if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
				consumeToken(parser);
			}

			VariableReassignNode *vrn = createVariableReassignNode();
			vrn->name = variableName;
			vrn->expr = expr;
			return vrn;
		}
	}

	printf(KRED "error: failed to parse variable reassignment\n" KNRM);
	exit(1);
}

void startParsingTokenStream(Parser *parser) {
	while (parser->parsing) {
		// get current token
		Token *tok = getItemFromVector(parser->tokenStream, parser->tokenIndex);

		switch (tok->type) {
			case IDENTIFIER:
				// parse a variable if we have a variable
				// given to us
				if (!strcmp(tok->content, "fn")) {
					parseFunctionNode(parser);
				} 
				else if (checkTokenTypeIsValidDataType(parser, tok)) {
					parseVariableNode(parser, true);
				}
				else if (checkTokenTypeAndContent(parser, OPERATOR, "=", 1)) {
					parseReassignmentStatementNode(parser);
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

bool checkTokenTypeIsValidDataType(Parser *parser, Token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return true;
		}
	}
	return false;
}

DataType matchTokenTypeToDataType(Parser *parser, Token *tok) {
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
	pushBackVectorItem(parser->parseTree, node);
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
			case FUNCTION_RET_NODE:
				destroyFunctionReturnNode(node->data);
				break;
			default:
				printf(KYEL "attempting to remove unrecognized node(%d)?\n" KNRM, node->type);
				break;
		}
	}
	free(node);
}

void destroyParser(Parser *parser) {
	int i;
	for (i = 0; i < parser->tokenStream->size; i++) {
		Token *tok = getItemFromVector(parser->tokenStream, i);
		destroyToken(tok);
	}
	destroyVector(parser->tokenStream);

	for (i = 0; i < parser->parseTree->size; i++) {
		Node *node = getItemFromVector(parser->parseTree, i);
		removeNode(node);
	}
	destroyVector(parser->parseTree);

	free(parser);
}