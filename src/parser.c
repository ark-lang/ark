#include "parser.h"

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void", "char"
};

static const char* NODE_TYPE[] = {
	"EXPRESSION_AST_NODE", "VARIABLE_DEF_AST_NODE",
	"VARIABLE_DEC_AST_NODE", "FUNCTION_ARG_AST_NODE",
	"FUNCTION_AST_NODE", "FUNCTION_PROT_AST_NODE",
	"BLOCK_AST_NODE", "FUNCTION_CALLEE_AST_NODE",
	"FUNCTION_RET_AST_NODE", "FOR_LOOP_AST_NODE",
	"VARIABLE_REASSIGN_AST_NODE", "INFINITE_LOOP_AST_NODE",
	"BREAK_AST_NODE", "DO_WHILE_AST_NODE", "CONTINUE_AST_NODE", "ENUM_AST_NODE", "STRUCT_AST_NODE",
	"IF_STATEMENT_AST_NODE", "MATCH_STATEMENT_AST_NODE", "WHILE_LOOP_AST_NODE",
	"ANON_AST_NODE", "USE_STATEMENT_AST_NODE"
};

/** Supported Operators */
static char* SUPPORTED_OPERANDS[] = {
	"++", "--",
	"+=", "-=", "*=", "/=", "%=",
	"+", "-", "&", "-", "*", "/", "%", "^", "**",
	">", "<", ">=", "<=", "==", "!=", "&&", "||",
};

/** UTILITY FOR AST NODES */

void exitParser(Parser *parser) {
	parser->parsing = false;
	parser->exitOnError = true;
}

void parserError(Parser *parser, char *msg, Token *tok, bool fatal_error) {
	errorMessage("%d:%d %s", tok->lineNumber, tok->charNumber, msg);
	char *error = getTokenContext(parser->tokenStream, tok);
	printf("\t%s\n", error);
	parser->exitOnError = fatal_error;
	free(error);
	if (fatal_error) {
		parser->parsing = false;
	}
}

void *allocateASTNode(size_t size, const char* readable_type) {
	// dont use safe malloc here because we can provide additional
	// error info
	void *ret = malloc(size);
	if (!ret) {
		fprintf(stderr, "malloc: failed to allocate memory for %s", readable_type);
		return NULL;
	}
	return ret;
}

FunctionOwnerAstNode *createFunctionOwnerAstNode() {
	FunctionOwnerAstNode *fo = allocateASTNode(sizeof(FunctionOwnerAstNode), "function owner");
	fo->owner = NULL;
	fo->alias = NULL;
	fo->isPointer = false;
	return fo;
}

InfiniteLoopAstNode *createInfiniteLoopAstNode() {
	InfiniteLoopAstNode *iln = allocateASTNode(sizeof(InfiniteLoopAstNode), "infinite loop");
	iln->body = NULL;
	return iln;
}

BreakStatementAstNode *createBreakStatementAstNode() {
	BreakStatementAstNode *bn = allocateASTNode(sizeof(BreakStatementAstNode), "break");
	return bn;
}

ContinueStatementAstNode *create_continue_ast_node() {
	ContinueStatementAstNode *cn = allocateASTNode(sizeof(ContinueStatementAstNode), "continue");
	return cn;
}

VariableReassignmentAstNode *createVariableReassignmentAstNode() {
	VariableReassignmentAstNode *vrn = allocateASTNode(sizeof(VariableReassignmentAstNode), "variable reassignment");
	vrn->name = NULL;
	vrn->expression = NULL;
	return vrn;
}

StatementAstNode *createStatementAstNode(void *data, AstNodeType type) {
	StatementAstNode *sn = allocateASTNode(sizeof(StatementAstNode), "statement");
	sn->data = data;
	sn->type = type;
	return sn;
}

FunctionReturnAstNode *createFunctionReturnAstNode() {
	FunctionReturnAstNode *frn = allocateASTNode(sizeof(FunctionReturnAstNode), "function return");
	frn->returnValue = NULL;
	return frn;
}

ExpressionAstNode *createExpressionAstNode() {
	ExpressionAstNode *expr = allocateASTNode(sizeof(ExpressionAstNode), "expression");
	expr->tokens = createVector();
	return expr;
}

VariableDefinitionAstNode *createVariableDefinitionAstNode() {
	VariableDefinitionAstNode *vdn = allocateASTNode(sizeof(VariableDefinitionAstNode), "variable definition");
	vdn->name = NULL;
	vdn->isGlobal = false;
	return vdn;
}

VariableDeclarationAstNode *createVariableDeclarationAstNode() {
	VariableDeclarationAstNode *vdn = allocateASTNode(sizeof(VariableDeclarationAstNode), "variable declaration");
	vdn->variableDefinitionAstNode = NULL;
	vdn->expression = NULL;
	return vdn;
}

FunctionArgumentAstNode *createFunctionArgumentAstNode() {
	FunctionArgumentAstNode *fan = allocateASTNode(sizeof(FunctionArgumentAstNode), "function argument");
	fan->name = NULL;
	fan->value = NULL;
	return fan;
}

FunctionCallAstNode *createFunctionCallAstNode() {
	FunctionCallAstNode *fcn = allocateASTNode(sizeof(FunctionCallAstNode), "function callee");
	fcn->name = NULL;
	fcn->args = NULL;
	fcn->isFunctionRedirect = false;
	return fcn;
}

BlockAstNode *createBlockAstNode() {
	BlockAstNode *bn = allocateASTNode(sizeof(BlockAstNode), "block");
	bn->statements = NULL;
	return bn;
}

FunctionPrototypeAstNode *createFunctionPrototypeAstNode() {
	FunctionPrototypeAstNode *fpn = allocateASTNode(sizeof(FunctionPrototypeAstNode), "function prototype");
	fpn->args = NULL;
	fpn->name = NULL;
	return fpn;
}

EnumAstNode *createEnumerationAstNode() {
	EnumAstNode *en = allocateASTNode(sizeof(EnumAstNode), "enum");
	en->name = NULL;
	en->enumItems = createVector();
	return en;
}

EnumeratedStructureAstNode *createEnumeratedStructureAstNode() {
	EnumeratedStructureAstNode *es = allocateASTNode(sizeof(EnumAstNode), "enum");
	es->name = NULL;
	es->structs = NULL;
	return es;
}

EnumItem *createEnumItem(char *name, int value) {
	EnumItem *ei = allocateASTNode(sizeof(EnumItem), "enum item");
	ei->name = name;
	ei->value = value;
	ei->hasValue = false;
	return ei;
}

FunctionAstNode *createFunctionAstNode() {
	FunctionAstNode *fn = allocateASTNode(sizeof(FunctionAstNode), "function");
	fn->prototype = NULL;
	fn->body = NULL;
	return fn;
}

ForLoopAstNode *createForLoopAstNode() {
	ForLoopAstNode *fln = allocateASTNode(sizeof(ForLoopAstNode), "for loop");
	return fln;
}

StructureAstNode *createStructureAstNode() {
	StructureAstNode *sn = allocateASTNode(sizeof(StructureAstNode), "struct");
	sn->statements = createVector();
	return sn;
}

UseStatementAstNode *createUseStatementAstNode() {
	UseStatementAstNode *useStatement = allocateASTNode(sizeof(UseStatementAstNode), "use");
	useStatement->file = NULL;
	return useStatement;
}

IfStatementAstNode *createIfStatementAstNode() {
	IfStatementAstNode *isn = allocateASTNode(sizeof(IfStatementAstNode), "if statement");
	return isn;
}

WhileLoopAstNode *createWhileLoopAstNode() {
	WhileLoopAstNode *wn = allocateASTNode(sizeof(WhileLoopAstNode), "while loop");
	return wn;
}

DoWhileAstNode *createDoWhileAstNode() {
	DoWhileAstNode *wn = allocateASTNode(sizeof(DoWhileAstNode), "do while loop");
	return wn;
}

MatchCaseAstNode *createMatchCaseAstNode() {
	MatchCaseAstNode *mcn = allocateASTNode(sizeof(MatchCaseAstNode), "match case");
	return mcn;
}

MatchAstNode *createMatchAstNode() {
	MatchAstNode *mn = allocateASTNode(sizeof(MatchAstNode), "match");
	mn->cases = createVector();
	return mn;
}

void destroyVariableReassignmentAstNode(VariableReassignmentAstNode *vrn) {
	if (vrn) {
		if (vrn->expression) {
			destroyExpressionAstNode(vrn->expression);
		}
		free(vrn);
	}
}

void destroyForLoopAstNode(ForLoopAstNode *loopNode) {
	if (loopNode) {
		destroyVector(loopNode->parameters);
		destroyBlockAstNode(loopNode->body);
		free(loopNode);
	}
}

void destroyFunctionOwnerAstNode(FunctionOwnerAstNode *functionOwner) {
	if (functionOwner) {
		free(functionOwner);
	}
}

void destroyBreakStatementAstNode(BreakStatementAstNode *breakStatement) {
	if (breakStatement) {
		free(breakStatement);
	}
}

void destroyContinueStatementAstNode(ContinueStatementAstNode *continueStatement) {
	if (continueStatement) {
		free(continueStatement);
	}
}

void destroyEnumeratedStructureAstNode(EnumeratedStructureAstNode *enumeratedStructure) {
	if (enumeratedStructure) {
		destroyVector(enumeratedStructure->structs);
		free(enumeratedStructure);
	}
}

void destroyStatementAstNode(StatementAstNode *statement) {
	if (statement) {
		if (statement->data) {
			switch (statement->type) {
				case VARIABLE_DEF_AST_NODE: destroyVariableDefinitionAstNode(statement->data); break;
				case VARIABLE_DEC_AST_NODE: destroyVariableDeclarationAstNode(statement->data); break;
				case FUNCTION_CALLEE_AST_NODE: destroyFunctionCallAstNode(statement->data); break;
				case FUNCTION_RET_AST_NODE: destroyFunctionAstNode(statement->data); break;
				case VARIABLE_REASSIGN_AST_NODE: destroyVariableReassignmentAstNode(statement->data); break;
				case FOR_LOOP_AST_NODE: destroyForLoopAstNode(statement->data); break;
				case INFINITE_LOOP_AST_NODE: destroyInfiniteLoopAstNode(statement->data); break;
				case BREAK_AST_NODE: destroyBreakStatementAstNode(statement->data); break;
				case CONTINUE_AST_NODE: destroyContinueStatementAstNode(statement->data); break;
				case ENUM_AST_NODE: destroyEnumAstNode(statement->data); break;
				case IF_STATEMENT_AST_NODE: destroyIfStatementAstNode(statement->data); break;
				case MATCH_STATEMENT_AST_NODE: destroyMatchAstNode(statement->data); break;
				case WHILE_LOOP_AST_NODE: destroyWhileLoopAstNode(statement->data); break;
				case DO_WHILE_AST_NODE: destroyDoWhileAstNode(statement->data); break;
				default: printf("trying to destroy unrecognized statement node %d\n", statement->type); break;
			}
		}
		free(statement);
	}
}

void destroyFunctionReturnAstNode(FunctionReturnAstNode *functionReturn) {
	if (functionReturn) {
		destroyExpressionAstNode(functionReturn->returnValue);
		free(functionReturn);
	}
}

void destroyExpressionAstNode(ExpressionAstNode *expression) {
	if (expression) {
		destroyVector(expression->tokens);
		free(expression);
	}
}

void destroyVariableDefinitionAstNode(VariableDefinitionAstNode *variableDefinition) {
	if (variableDefinition) {
		free(variableDefinition);
	}
}

void destroyVariableDeclarationAstNode(VariableDeclarationAstNode *variableDeclaration) {
	if (variableDeclaration) {
		destroyVariableDefinitionAstNode(variableDeclaration->variableDefinitionAstNode);
		destroyExpressionAstNode(variableDeclaration->expression);
		free(variableDeclaration);
	}
}

void destroyFunctionArgumentAstNode(FunctionArgumentAstNode *functionArgument) {
	if (functionArgument) {
		if (functionArgument->value) {
			destroyExpressionAstNode(functionArgument->value);
		}
		free(functionArgument);
	}
}

void destroyBlockAstNode(BlockAstNode *blockNode) {
	if (blockNode) {
		if (blockNode->statements) {
			destroyVector(blockNode->statements);
		}
		free(blockNode);
	}
}

void destroyInfiniteLoopAstNode(InfiniteLoopAstNode *infiniteLoop) {
	if (infiniteLoop) {
		if (infiniteLoop->body) {
			destroyBlockAstNode(infiniteLoop->body);
		}
		free(infiniteLoop);
	}
}

void destroyFunctionPrototypeAstNode(FunctionPrototypeAstNode *functionPrototype) {
	if (functionPrototype) {
		if (functionPrototype->args) {
			int i;
			for (i = 0; i < functionPrototype->args->size; i++) {
				StatementAstNode *sn = getVectorItem(functionPrototype->args, i);
				if (sn) {
					destroyStatementAstNode(sn);
				}
			}
			destroyVector(functionPrototype->args);
		}
		free(functionPrototype);
	}
}

void destroyFunctionAstNode(FunctionAstNode *function) {
	if (function) {
		if (function->prototype) {
			destroyFunctionPrototypeAstNode(function->prototype);
		}
		if (function->body) {
			destroyBlockAstNode(function->body);
		}
		free(function);
	}
}

void destroyFunctionCallAstNode(FunctionCallAstNode *functionCall) {
	if (functionCall) {
		if (functionCall->args) {
			destroyVector(functionCall->args);
		}
		free(functionCall);
	}
}

void destroyUseStatementAstNode(UseStatementAstNode *useStatement) {
	if (useStatement) {
		free(useStatement);
	}
}

void destroyStructureAstNode(StructureAstNode *structure) {
	if (structure) {
		if (structure->statements) {
			destroyVector(structure->statements);
		}
		free(structure);
	}
}

void destroyEnumAstNode(EnumAstNode *enumeration) {
	if (enumeration) {
		if (enumeration->enumItems) {
			int i;
			for (i = 0; i < enumeration->enumItems->size; i++) {
				destroyEnumItem(getVectorItem(enumeration->enumItems, i));
			}
			destroyVector(enumeration->enumItems);
		}
		free(enumeration);
	}
}

void destroyIfStatementAstNode(IfStatementAstNode *ifStatement) {
	if (ifStatement) {
		if (ifStatement->condition) {
			destroyExpressionAstNode(ifStatement->condition);
		}
		if (ifStatement->body) {
			destroyBlockAstNode(ifStatement->body);
		}
		free(ifStatement);
	}
}

void destroyWhileLoopAstNode(WhileLoopAstNode *whileLoop) {
	if (whileLoop) {
		if (whileLoop->condition) {
			destroyExpressionAstNode(whileLoop->condition);
		}
		if (whileLoop->body) {
			destroyBlockAstNode(whileLoop->body);
		}
		free(whileLoop);
	}
}

void destroyDoWhileAstNode(DoWhileAstNode *doWhile) {
	if (doWhile) {
		if (doWhile->condition) {
			destroyExpressionAstNode(doWhile->condition);
		}
		if (doWhile->body) {
			destroyBlockAstNode(doWhile->body);
		}
		free(doWhile);
	}
}


void destroyMastCaseAstNode(MatchCaseAstNode *matchCase) {
	if (matchCase) {
		if (matchCase->condition) {
			destroyExpressionAstNode(matchCase->condition);
		}
		if (matchCase->body) {
			destroyBlockAstNode(matchCase->body);
		}
		free(matchCase);
	}
}

void destroyMatchAstNode(MatchAstNode *match) {
	if (match) {
		if (match->condition) {
			destroyExpressionAstNode(match->condition);
		}
		if (match->cases) {
			int i;
			for (i = 0; i < match->cases->size; i++) {
				destroyMastCaseAstNode(getVectorItem(match->cases, i));
			}
			destroyVector(match->cases);
		}
		free(match);
	}
}

void destroyEnumItem(EnumItem *enumItem) {
	if (enumItem) {
		free(enumItem);
	}
}

/** END AST_NODE FUNCTIONS */

Parser *createParser() {
	Parser *parser = safeMalloc(sizeof(*parser));
	parser->tokenStream = NULL;
	parser->tokenIndex = 0;
	parser->parsing = true;
	parser->exitOnError = false;
	return parser;
}

Token *consumeToken(Parser *parser) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex++);
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex + ahead);
}

Token *expectTokenType(Parser *parser, TokenType type) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (tok->type == type) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unrecognized token found", tok, true);
		return NULL;
	}
}

Token *expectTokenContent(Parser *parser, char *content) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (!strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

Token *expectTokenTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = peekAtTokenStream(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

Token *matchTokenType(Parser *parser, TokenType type) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (tok->type == type) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

Token *matchTokenContent(Parser *parser, char *content) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (!strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

Token *matchTokenTypeAndContent(Parser *parser, TokenType type, char *content) {
	Token *tok = peekAtTokenStream(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consumeToken(parser);
	}
	else {
		parserError(parser, "unexpected token found", tok, true);
		return NULL;
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

int parseOperand(Parser *parser) {
	Token *token = peekAtTokenStream(parser, 0);

	int i;
	int operandListSize = sizeof(SUPPORTED_OPERANDS) / sizeof(SUPPORTED_OPERANDS[0]);
	for (i = 0; i < operandListSize; i++) {
		if (!strcmp(SUPPORTED_OPERANDS[i], token->content)) {
			consumeToken(parser);
			return i;
		}
	}

	parserError(parser, "unsupported operator specified", token, true);
	return -1;
}

WhileLoopAstNode *parseWhileAstNode(Parser *parser) {
	WhileLoopAstNode *whileLoop = createWhileLoopAstNode();
	matchTokenTypeAndContent(parser, IDENTIFIER, WHILE_LOOP_KEYWORD);

	whileLoop->condition = parseExpressionAstNode(parser);
	whileLoop->body = parseBlockAstNode(parser);

	return whileLoop;
}

IfStatementAstNode *parseIfStatementAstNode(Parser *parser) {
	IfStatementAstNode *ifStatement = createIfStatementAstNode();

	matchTokenTypeAndContent(parser, IDENTIFIER, IF_KEYWORD);

	ifStatement->condition = parseExpressionAstNode(parser);
	ifStatement->body = parseBlockAstNode(parser);
	ifStatement->statementType = IF_STATEMENT;

	if (checkTokenTypeAndContent(parser, IDENTIFIER, ELSE_KEYWORD, 0)) {
		consumeToken(parser);
		ifStatement->statementType = ELSE_STATEMENT;
		ifStatement->elseStatement = parseBlockAstNode(parser);
	}

	return ifStatement;
}

MatchCaseAstNode *parseMatchCaseAstNode(Parser *parser) {
	MatchCaseAstNode *matchCase = createMatchCaseAstNode();

	matchCase->condition = parseExpressionAstNode(parser);
	matchCase->body = parseBlockAstNode(parser);

	return matchCase;
}

MatchAstNode *parseMatchAstNode(Parser *parser) {
	MatchAstNode *match = createMatchAstNode();

	matchTokenTypeAndContent(parser, IDENTIFIER, MATCH_KEYWORD);

	ExpressionAstNode *expr = parseExpressionAstNode(parser);
	match->condition = expr;
	match->cases = createVector();

	if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consumeToken(parser);

		do {
			if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consumeToken(parser);
				break;
			}

			pushBackItem(match->cases, parseMatchCaseAstNode(parser));
		}
		while (true);
	}
	else {
		parserError(parser, "match expected block denoted with `{}`", consumeToken(parser), true);
	}

	return match;
}

EnumAstNode *parseEnumerationAstNode(Parser *parser) {
	EnumAstNode *enumeration = createEnumerationAstNode();

	matchTokenTypeAndContent(parser, IDENTIFIER, ENUM_KEYWORD); // ENUM

	// ENUMERATIONS NAME
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *enumName = consumeToken(parser);
		enumeration->name = enumName;

		// OPEN OF ENUM BLOCK
		if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_OPENER, 0)) {
			consumeToken(parser);

			// LOOP
			do {

				// eat the last brace
				if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
					consumeToken(parser);
					break;
				}

				if (checkTokenType(parser, IDENTIFIER, 0)) {
					Token *enumItemName = consumeToken(parser);

					// setting the enum = to a value
					if (checkTokenTypeAndContent(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
						consumeToken(parser);

						// to a number
						if (checkTokenType(parser, NUMBER, 0)) {
							Token *enumItemValue = consumeToken(parser);

							// convert to int
							int enumItemValueAsInt = atoi(enumItemValue->content);

							// if we already have items in our enum
							// make sure there are no duplicate values or names
							if (enumeration->enumItems->size >= 1) {
								EnumItem *previousEnumItem = getVectorItem(enumeration->enumItems, enumeration->enumItems->size - 1);
								int previousEnumItemValue = previousEnumItem->value;
								char *previousEnumItemName = previousEnumItem->name;

								// validate names are not duplicate
								if (!strcmp(previousEnumItemName, enumItemName->content)) {
									parserError(parser, "duplicate item in enumeration", enumItemName, false);
								}

								// validate values are not duplicate
								if (previousEnumItemValue == enumItemValueAsInt) {
									parserError(parser, "duplicate item value in enumeration", enumItemName, false);
								}
							}

							// push it back
							EnumItem *enumItem = createEnumItem(enumItemName->content, enumItemValueAsInt);
							enumItem->hasValue = true;
							pushBackItem(enumeration->enumItems, enumItem);
						}
						else {
							parserError(parser, "invalid integer literal assigned to enumeration item", consumeToken(parser), false);
						}

					}
					// ENUM_ITEM with no assignment
					else {
						int enumItemValueAsInt = 0;

						if (enumeration->enumItems->size >= 1) {
							EnumItem *previousEnumItem = getVectorItem(enumeration->enumItems, enumeration->enumItems->size - 1);
							enumItemValueAsInt = previousEnumItem->value + 1;
							char *previousEnumItemName = previousEnumItem->name;

							// validate name
							if (!strcmp(previousEnumItemName, enumItemName->content)) {
								parserError(parser, "duplicate item in enumeration", enumItemName, false);
							}
						}

						EnumItem *enumItem = createEnumItem(enumItemName->content, enumItemValueAsInt);
						pushBackItem(enumeration->enumItems, enumItem);
					}
				}

				if (checkTokenTypeAndContent(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
					consumeToken(parser);
					if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
						parserError(parser, "trailing comma in enumeration", consumeToken(parser), false);
						break;
					}
				}

				// we've finished parsing jump out
				if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
					consumeToken(parser);
					break;
				}

			}
			while (true);

			// empty enum, throw an error.
			if (enumeration->enumItems->size == 0) {
				parserError(parser, "empty enumeration", consumeToken(parser), false);
			}
		}
	}

	return enumeration;
}

EnumeratedStructureAstNode *parseEnumeratedStructureAstNode(Parser *parser) {
	EnumeratedStructureAstNode *enumeratedStructure = createEnumeratedStructureAstNode();

	matchTokenTypeAndContent(parser, IDENTIFIER, ANON_STRUCT_KEYWORD);
	Token *estruct_name = matchTokenType(parser, IDENTIFIER);
	enumeratedStructure->name = estruct_name;
	enumeratedStructure->structs = createVector();

	if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consumeToken(parser);
		do {
			if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consumeToken(parser);
				break;
			}
			pushBackItem(enumeratedStructure->structs, matchTokenType(parser, IDENTIFIER));
		}
		while (true);
	}

	return enumeratedStructure;
}

StructureAstNode *parseStructureAstNode(Parser *parser) {
	matchTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD);
	Token *structName = matchTokenType(parser, IDENTIFIER);

	StructureAstNode *structure = createStructureAstNode();
	structure->name = structName->content;

	// parses a block of statements
	if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consumeToken(parser);

		do {
			if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consumeToken(parser);
				break;
			}

			pushBackItem(structure->statements, parseVariableAstNode(parser, false));
		}
		while (true);
	}

	return structure;
}

StatementAstNode *parseForLoopAstNode(Parser *parser) {
	// create node with the stuff we just got
	ForLoopAstNode *forLoop = createForLoopAstNode();
	forLoop->parameters = createVector();

	// for token
	matchTokenTypeAndContent(parser, IDENTIFIER, FOR_LOOP_KEYWORD);			// FOR KEYWORD

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		forLoop->type = consumeToken(parser);
	}
	else {
		parserError(parser, "Expected a type in For Loop", consumeToken(parser), false);
	}

	Token *indexName = matchTokenType(parser, IDENTIFIER);					// INDEX_NAME
	forLoop->indexName = indexName;

	matchTokenTypeAndContent(parser, OPERATOR, ":");						// PARAMS

	// consume the args
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		Token *argumentOpener = consumeToken(parser);

		int paramCount = 0;

		do {
			if (paramCount > MAX_FOR_LOOP_PARAM_COUNT) {
				parserError(parser, "too many parameters passed to for loop", consumeToken(parser), false);
			}
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				if (paramCount < MIN_FOR_LOOP_PARAM_COUNT) {
					parserError(parser, "too few parameters passed to for loop", argumentOpener, false);
				}
				consumeToken(parser);
				break;
			}

			if (checkTokenType(parser, IDENTIFIER, 0)) {
				Token *token = consumeToken(parser);
				pushBackItem(forLoop->parameters, token);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						parserError(parser, "trailing comma in for loop declaration", consumeToken(parser), false);
					}
					consumeToken(parser);
				}
			}
			else if (checkTokenType(parser, NUMBER, 0)) {
				Token *token = consumeToken(parser);
				pushBackItem(forLoop->parameters, token);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						parserError(parser, "trailing comma in for loop declaration", consumeToken(parser), false);
					}
					consumeToken(parser);
				}
			}
			// it's an expression probably
			else if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
				ExpressionAstNode *expr = parseExpressionAstNode(parser);
				pushBackItem(forLoop->parameters, expr);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						parserError(parser, "trailing comma in for loop declaration", consumeToken(parser), false);
					}
					consumeToken(parser);
				}
			}
			else {
				parserError(parser, "expected a number literal or a variable in for loop parameters", consumeToken(parser), false);
				break;
			}

			paramCount++;
		}
		while (true);

		forLoop->body = parseBlockAstNode(parser);

		return createStatementAstNode(forLoop, FOR_LOOP_AST_NODE);
	}

	parserError(parser, "failed to parse for loop", consumeToken(parser), true);
	return NULL;
}

ExpressionAstNode *parseExpressionAstNode(Parser *parser) {
	ExpressionAstNode *expr = createExpressionAstNode();

	while (true) {
		// token is separator, semi colon, etc
		if (checkTokenContent(parser, "^", 0)) {
			Token *tok = consumeToken(parser);
			tok->content[0] = '*';
			pushBackItem(expr->tokens, tok);
		}
		else {
			pushBackItem(expr->tokens, consumeToken(parser));
		}
		if (checkTokenContent(parser, "{", 0)
			|| checkTokenContent(parser, ",", 0)
			|| checkTokenContent(parser, ";", 0)
			|| (checkTokenContent(parser, "->", 0))
			|| (checkTokenContent(parser, ")", 0) && checkTokenContent(parser, ";", 1))
			|| (checkTokenContent(parser, ")", 0) && checkTokenContent(parser, "->", 1))) {
			return expr;
		}
	}

	return expr;
}

void *parseVariableAstNode(Parser *parser, bool isGlobal) {
	bool isMutable = false;

	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		isMutable = true;
	}

	// consume the int data type
	Token *variableDataType = matchTokenType(parser, IDENTIFIER);
	VariableDefinitionAstNode *def = createVariableDefinitionAstNode();

	// is a pointer
	bool isPointer = false;
	if (checkTokenTypeAndContent(parser, OPERATOR, POINTER_OPERATOR, 0)) {
		isPointer = true;
		consumeToken(parser);
	}

	// name of the variable
	Token *variableNameToken = matchTokenType(parser, IDENTIFIER);

	if (checkTokenTypeAndContent(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
		// consume the equals sign
		consumeToken(parser);

		// create variable define ast_node
		def->type = variableDataType;
		def->name = variableNameToken->content;
		def->isGlobal = isGlobal;
		def->isPointer = isPointer;

		// create the variable declare ast_node
		VariableDeclarationAstNode *dec = createVariableDeclarationAstNode();
		dec->isMutable = isMutable;
		dec->variableDefinitionAstNode = def;
		dec->expression = parseExpressionAstNode(parser);

		// this is weird, we can probably clean this up
		if (isGlobal) {
			pushAstNode(parser, dec, VARIABLE_DEC_AST_NODE);
			return dec;
		}

		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		else {
			parserError(parser, "Expected a semi-colon at the end of variable declaration", consumeToken(parser), false);
		}

		// not global, pop it as a statement node
		return createStatementAstNode(dec, VARIABLE_DEC_AST_NODE);
	}
	else {
		// create variable define ast_node
		def->type = variableDataType;
		def->name = variableNameToken->content;
		def->isGlobal = isGlobal;
		def->isPointer = isPointer;

		if (isGlobal) {
			pushAstNode(parser, def, VARIABLE_DEF_AST_NODE);
			return def;
		}

		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		else {
			parserError(parser, "Expected a semi-colon at the end of variable definition", consumeToken(parser), false);
		}

		// not global, pop it as a statement node
		return createStatementAstNode(def, VARIABLE_DEF_AST_NODE);
	}
}

BlockAstNode *parseBlockAstNode(Parser *parser) {
	BlockAstNode *block = createBlockAstNode();
	block->statements = createVector();
	block->isSingleStatement = false;

	if (checkTokenTypeAndContent(parser, OPERATOR, SINGLE_STATEMENT, 0)) {
		consumeToken(parser);
		pushBackItem(block->statements, parseStatementAstNode(parser));
		block->isSingleStatement = true;
	}
	else if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consumeToken(parser);

		do {
			// check if block is empty before we try parse some statements
			if (checkTokenTypeAndContent(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consumeToken(parser);
				break;
			}

			pushBackItem(block->statements, parseStatementAstNode(parser));
		}
		while (true);
	}
	else {
		parserError(parser, "expected a multi-block or single-block statement", consumeToken(parser), true);
	}

	return block;
}

StatementAstNode *parseInfiniteLoopAstNode(Parser *parser) {
	matchTokenType(parser, IDENTIFIER);

	BlockAstNode *body = parseBlockAstNode(parser);

	InfiniteLoopAstNode *infiniteLoop = createInfiniteLoopAstNode();
	infiniteLoop->body = body;

	return createStatementAstNode(infiniteLoop, INFINITE_LOOP_AST_NODE);
}

UseStatementAstNode *parseUseStatementAstNode(Parser *parser) {
	UseStatementAstNode *use = createUseStatementAstNode();
	matchTokenTypeAndContent(parser, IDENTIFIER, USE_KEYWORD);

	if (checkTokenType(parser, STRING, 0)) {
		use->file = consumeToken(parser);
	}
	else {
		parserError(parser, "Expected a file to use", consumeToken(parser), true);
	}

	pushAstNode(parser, use, USE_STATEMENT_AST_NODE);
	return use;
}

FunctionAstNode *parseFunctionAstNode(Parser *parser) {
	matchTokenType(parser, IDENTIFIER);	// consume the fn keyword

	FunctionOwnerAstNode *functionOwner = NULL;

	// we're specifying an owner, so it's a method!
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		functionOwner = createFunctionOwnerAstNode();
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			functionOwner->owner = consumeToken(parser);

			// check if the owner is a pointer or not
			if (checkTokenTypeAndContent(parser, OPERATOR, POINTER_OPERATOR, 0)) {
				functionOwner->isPointer = true;
			}

			// check if we give an owner
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				functionOwner->alias = consumeToken(parser);
			}
			else {
				parserError(parser, "method expected identifier to specify an alias for the method owner", consumeToken(parser), false);
				destroyFunctionOwnerAstNode(functionOwner);
			}

			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
			}
		}
		else {
			parserError(parser, "method expected an identifier to specify ownership", consumeToken(parser), false);
			destroyFunctionOwnerAstNode(functionOwner);
		}
	}

	Token *functionName = matchTokenType(parser, IDENTIFIER); // name of function
	Vector *args = createVector();

	// Create function signature
	FunctionPrototypeAstNode *prototype = createFunctionPrototypeAstNode();
	prototype->args = args;
	prototype->name = functionName;
	prototype->owner = functionOwner;

	// parameter list
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			bool isMutable = false;
			if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
				isMutable = true;
				consumeToken(parser);
			}

			// data type
			Token *argumentType = matchTokenType(parser, IDENTIFIER);

			// look for ^
			bool isPointer = false;
			if (checkTokenTypeAndContent(parser, OPERATOR, POINTER_OPERATOR, 0)) {
				isPointer = true;
				consumeToken(parser);
			}

			// name of argument
			Token *argumentName = matchTokenType(parser, IDENTIFIER);

			FunctionArgumentAstNode *arg = createFunctionArgumentAstNode();
			arg->type = argumentType;
			arg->name = argumentName;
			arg->isPointer = isPointer;
			arg->isMutable = isMutable;
			arg->value = NULL;

			if (checkTokenTypeAndContent(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
				consumeToken(parser);

				// default expression
				ExpressionAstNode *expr = parseExpressionAstNode(parser);
				arg->value = expr;
				pushBackItem(args, arg);

				if (checkTokenTypeAndContent(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
					consumeToken(parser);
				}
				else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser); // eat closing parenthesis
					break;
				}
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					parserError(parser, "trailing comma at the end of argument list", consumeToken(parser), false);
				}
				consumeToken(parser); // eat the comma
				pushBackItem(args, arg);
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser); // eat closing parenthesis
				pushBackItem(args, arg);
				break;
			}
		}
		while (true);

		FunctionAstNode *function = createFunctionAstNode();

		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);

			bool isMutable = false;
			if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
				isMutable = true;
				consumeToken(parser);
			}

			bool returnPointer = false;
			if (checkTokenTypeAndContent(parser, OPERATOR, POINTER_OPERATOR, 0)) {
				returnPointer = true;
				consumeToken(parser);
			}

			// returns data type
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				Token *returnType = consumeToken(parser);
				prototype->returnType = returnType;
				function->returnsPointer = returnPointer;
				function->isMutable = isMutable;
			}
			else {
				parserError(parser, "function declaration return type expected", consumeToken(parser), false);
			}
		}
		else if (checkTokenType(parser, IDENTIFIER, 0)) {
			// if they do for example
			//              V forgot the :!!!!
			// fn whatever() int {
			// }
			parserError(parser, "found an identifier after function argument list, perhaps you missed a colon?", consumeToken(parser), false);
		}
		else {
			prototype->returnType = NULL;
		}

		// set function prototype
		function->prototype = prototype;
		function->body = parseBlockAstNode(parser);

		pushAstNode(parser, function, FUNCTION_AST_NODE);

		return function;
	}
	else {
		parserError(parser, "expecting a parameter list", consumeToken(parser), false);
	}

	// just in case we fail to parse, free this shit
	free(prototype);
	prototype = NULL;

	parserError(parser, "failed to parse function", consumeToken(parser), true);
	return NULL;
}

FunctionCallAstNode *parseFunctionCallAstNode(Parser *parser) {
	// consume function name
	Token *call = matchTokenType(parser, IDENTIFIER);
	FunctionCallAstNode *functionCall = createFunctionCallAstNode();

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);	// eat open bracket

		Vector *args = createVector();

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			ExpressionAstNode *expression = parseExpressionAstNode(parser);

			FunctionArgumentAstNode *argument = createFunctionArgumentAstNode();
			argument->value = expression;

			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					parserError(parser, "trailing comma at the end of argument list", consumeToken(parser), false);
				}
				consumeToken(parser);
				pushBackItem(args, argument);
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser); // eat closing parenthesis
				pushBackItem(args, argument);
				break;
			}
		}
		while (true);

		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		else if (checkTokenContent(parser, "->", 0)) {
			consumeToken(parser);
			Vector *identifiers = createVector();

			while (true) {
				pushBackItem(identifiers, consumeToken(parser));

				if (checkTokenContent(parser, ",", 0)) {
					if (checkTokenContent(parser, ";", 1)) {
						parserError(parser, "Error, trailing comma in function redirect!", consumeToken(parser), false);
					}
					consumeToken(parser);
				}
				else {
					if (checkTokenContent(parser, ";", 0)) {
						consumeToken(parser);
						break;
					}
				}
			}

			functionCall->vars = identifiers;
			functionCall->isFunctionRedirect = true;
		}
		else {
			parserError(parser, "Expected a semi-colon at the end of function call", consumeToken(parser), true);
		}

		functionCall->name = call->content;
		functionCall->args = args;

		return functionCall;
	}

	destroyFunctionCallAstNode(functionCall);
	parserError(parser, "failed to parse function call", consumeToken(parser), true);
	return NULL;
}

FunctionReturnAstNode *parseReturnStatementAstNode(Parser *parser) {
	// consume the return keyword
	matchTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD);

	// return value
	FunctionReturnAstNode *functionReturn = createFunctionReturnAstNode();
	functionReturn->returnValue = parseExpressionAstNode(parser);

	if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		consumeToken(parser);
	}
	else {
		parserError(parser, "Expected a semi-colon at the end of return statement", consumeToken(parser), true);
	}

	return functionReturn;

	parserError(parser, "failed to parse return statement", consumeToken(parser), true);
	return NULL;
}

BreakStatementAstNode *parseBreakStatementAstNode(Parser *parser) {
	BreakStatementAstNode *breakStmt = createBreakStatementAstNode();
	matchTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD);
	// eat dat semi colon bby gurl
	if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		consumeToken(parser);
	}
	return breakStmt;
}

DoWhileAstNode *parseDoWhileAstNode(Parser *parser) {
	DoWhileAstNode *doWhile = createDoWhileAstNode();
	matchTokenTypeAndContent(parser, IDENTIFIER, DO_KEYWORD);

	doWhile->condition = parseExpressionAstNode(parser);
	doWhile->body = parseBlockAstNode(parser);

	return doWhile;
}

StatementAstNode *parseStatementAstNode(Parser *parser) {
	// RETURN STATEMENTS
	if (checkTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		return createStatementAstNode(parseReturnStatementAstNode(parser), FUNCTION_RET_AST_NODE);
	}
	// BREAK STATEMENT
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		return createStatementAstNode(parseBreakStatementAstNode(parser), BREAK_AST_NODE);
	}
	// DO WHILE LOOP
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, DO_KEYWORD, 0)) {
		return createStatementAstNode(parseDoWhileAstNode(parser), DO_WHILE_AST_NODE);
	}
	// STRUCTURES
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		return createStatementAstNode(parseStructureAstNode(parser), STRUCT_AST_NODE);
	}
	// IF STATEMENTS
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, IF_KEYWORD, 0)) {
		return createStatementAstNode(parseIfStatementAstNode(parser), IF_STATEMENT_AST_NODE);
	}
	// MATCH STATEMENTS
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, MATCH_KEYWORD, 0)) {
		return createStatementAstNode(parseMatchAstNode(parser), MATCH_STATEMENT_AST_NODE);
	}
	// WHILE LOOPS
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, WHILE_LOOP_KEYWORD, 0)) {
		return createStatementAstNode(parseWhileAstNode(parser), WHILE_LOOP_AST_NODE);
	}
	// FOR LOOPS
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, FOR_LOOP_KEYWORD, 0)) {
		return parseForLoopAstNode(parser);
	}
	// INFINITE LOOPS
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, INFINITE_LOOP_KEYWORD, 0)) {
		return parseInfiniteLoopAstNode(parser);
	}
	// ENUMERATION
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, ENUM_KEYWORD, 0)) {
		return createStatementAstNode(parseEnumerationAstNode(parser), ENUM_AST_NODE);
	}
	// IDENTIFERS
	else if (checkTokenType(parser, IDENTIFIER, 0)) {
		// VARIABLE REASSIGNMENT
		if (checkTokenTypeAndContent(parser, OPERATOR, ASSIGNMENT_OPERATOR, 1)) {
			return createStatementAstNode(parseReassignmentAstNode(parser), VARIABLE_REASSIGN_AST_NODE);
		}
		// FUNCITON CALL
		else if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 1)) {
			return createStatementAstNode(parseFunctionCallAstNode(parser), FUNCTION_CALLEE_AST_NODE);
		}
		// no clue we should sort this out.
		else {
			return parseVariableAstNode(parser, false);
		}
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, "^", 0)) {
		if (checkTokenTypeAndContent(parser, OPERATOR, ASSIGNMENT_OPERATOR, 2)) {
			return createStatementAstNode(parseReassignmentAstNode(parser), VARIABLE_REASSIGN_AST_NODE);
		}
	}

	parserError(parser, "unrecognized token specified", consumeToken(parser), true);
	return NULL;
}

VariableReassignmentAstNode *parseReassignmentAstNode(Parser *parser) {
	bool isPointer = false;
	if (checkTokenTypeAndContent(parser, OPERATOR, "^", 0)) {
		consumeToken(parser);
		isPointer = true;
	}

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *variableName = consumeToken(parser);

		if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
			consumeToken(parser);

			ExpressionAstNode *expr = parseExpressionAstNode(parser);

			if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
				consumeToken(parser);
			}
			else {
				parserError(parser, "Expected a semi-colon at the end of return statement", consumeToken(parser), true);
			}

			VariableReassignmentAstNode *reassignment = createVariableReassignmentAstNode();
			reassignment->name = variableName;
			reassignment->expression = expr;
			reassignment->isPointer = isPointer;
			return reassignment;
		}
	}

	parserError(parser, "failed to parse variable reassignment", consumeToken(parser), true);
	return NULL;
}

void startParsingSourceFiles(Parser *parser, Vector *sourceFiles) {
	int i;
	for (i = 0; i < sourceFiles->size; i++) {
		SourceFile *file = getVectorItem(sourceFiles, i);
		parser->tokenStream = file->tokens;
		parser->parseTree = createVector();
		parser->tokenIndex = 0;
		parser->parsing = true;

		parseTokenStream(parser);

		file->ast = parser->parseTree;
	}
}

void parseTokenStream(Parser *parser) {
	while (parser->parsing) {
		// get current token
		Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex);

		// clean this too
		switch (tok->type) {
			case IDENTIFIER:
				if (!strcmp(tok->content, USE_KEYWORD)) {
					parseUseStatementAstNode(parser);
				}
				else if (!strcmp(tok->content, FUNCTION_KEYWORD)) {
					parseFunctionAstNode(parser);
				}
				else if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
					pushAstNode(parser, parseStructureAstNode(parser), STRUCT_AST_NODE);
				}
				else if (checkTokenTypeAndContent(parser, IDENTIFIER, ANON_STRUCT_KEYWORD, 0)) {
					pushAstNode(parser, parseEnumeratedStructureAstNode(parser), ANON_AST_NODE);
				}
				else if (checkTokenTypeAndContent(parser, IDENTIFIER, ENUM_KEYWORD, 0)) {
					pushAstNode(parser, parseEnumerationAstNode(parser), ENUM_AST_NODE);
				}
				else if (checkTokenTypeAndContent(parser, OPERATOR, "=", 1)) {
					parseReassignmentAstNode(parser);
				}
				else {
					parseVariableAstNode(parser, true);
				}
				break;
			case END_OF_FILE:
				parser->parsing = false;
				break;
		}
	}
}

bool validateTokenType(Parser *parser, Token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return true;
		}
	}
	return false;
}

void pushAstNode(Parser *parser, void *data, AstNodeType type) {
	AstNode *astNode = safeMalloc(sizeof(*astNode));
	astNode->data = data;
	astNode->type = type;
	pushBackItem(parser->parseTree, astNode);
}

void removeAstNode(AstNode *astNode) {
	/**
	 * This could probably be a lot more cleaner
	 */
	if (!astNode->data) {
		switch (astNode->type) {
			case EXPRESSION_AST_NODE:
				destroyExpressionAstNode(astNode->data);
				break;
			case VARIABLE_DEF_AST_NODE:
				destroyVariableDefinitionAstNode(astNode->data);
				break;
			case USE_STATEMENT_AST_NODE:
				destroyUseStatementAstNode(astNode->data);
				break;
			case VARIABLE_DEC_AST_NODE:
				destroyVariableDeclarationAstNode(astNode->data);
				break;
			case FUNCTION_ARG_AST_NODE:
				destroyFunctionArgumentAstNode(astNode->data);
				break;
			case FUNCTION_AST_NODE:
				destroyFunctionAstNode(astNode->data);
				break;
			case FUNCTION_PROT_AST_NODE:
				destroyFunctionPrototypeAstNode(astNode->data);
				break;
			case BLOCK_AST_NODE:
				destroyBlockAstNode(astNode->data);
				break;
			case ANON_AST_NODE:
				destroyEnumeratedStructureAstNode(astNode->data);
				break;
			case FUNCTION_CALLEE_AST_NODE:
				destroyFunctionCallAstNode(astNode->data);
				break;
			case FUNCTION_RET_AST_NODE:
				destroyFunctionReturnAstNode(astNode->data);
				break;
			case FOR_LOOP_AST_NODE:
				destroyForLoopAstNode(astNode->data);
				break;
			case VARIABLE_REASSIGN_AST_NODE:
				destroyVariableReassignmentAstNode(astNode->data);
				break;
			case INFINITE_LOOP_AST_NODE:
				destroyInfiniteLoopAstNode(astNode->data);
				break;
			case BREAK_AST_NODE:
				destroyBreakStatementAstNode(astNode->data);
				break;
			case ENUM_AST_NODE:
				destroyEnumAstNode(astNode->data);
				break;
			default:
				errorMessage("attempting to remove unrecognized ast_node(%d)?\n", astNode->type);
				break;
		}
	}
	free(astNode);
}

void destroyParser(Parser *parser) {
	if (parser) {
		int i;
		for (i = 0; i < parser->parseTree->size; i++) {
			AstNode *astNode = getVectorItem(parser->parseTree, i);
			debugMessage("Removed node %s", NODE_TYPE[astNode->type]);
			removeAstNode(astNode);
		}
		destroyVector(parser->parseTree);

		free(parser);
		debugMessage("Destroyed parser");
	}
}
