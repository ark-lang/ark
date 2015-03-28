#include "parser.h"

Parser *createParser() {
	Parser *parser = safeMalloc(sizeof(*parser));
	parser->tokenStream = NULL;
	parser->tokenIndex = 0;
	parser->parsing = true;
	parser->failed = false;
	return parser;
}

void destroyParser(Parser *parser) {
	free(parser);
	debugMessage("Destroyed parser");
}

/** PARSING STUFF */

Literal *parseLiteral(Parser *parser) {
	Token *tok = consumeToken(parser);
	LiteralType type = getLiteralType(tok);
	if (type == LITERAL_ERRORED) {
		errorMessage("Errored at literal");
		return NULL;
	}
	return createLiteral(tok->content, type);
}

IdentifierList *parseIdentifierList(Parser *parser) {
	IdentifierList *idenList = createIdentifierList();

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		pushBackItem(idenList->values, consumeToken(parser)->content);

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				consumeToken(parser);
			}
			else {
				errorMessage("Expected a comma, but found `%s`", consumeToken(parser)->content);
				break; // should go down to the error below
			}

			if (checkTokenType(parser, IDENTIFIER, 0)) {
				// no more commas, exit out of list!
				if (!checkTokenTypeAndContent(parser, SEPARATOR, ",", 1)) {
					pushBackItem(idenList->values, consumeToken(parser)->content);
					return idenList;
				}
				pushBackItem(idenList->values, consumeToken(parser)->content);
			}
		}
	}
	else {
		errorMessage("Expected an identifier, but found `%s`", consumeToken(parser)->content);
	}

	destroyIdentifierList(idenList);
	return NULL;
}

Type *parseType(Parser *parser) {
	Type *type = createType();

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *token = consumeToken(parser);
		type->typeName = createTypeName(token->content);
		type->type = TYPE_NAME_NODE;
	}
	else if (checkTokenTypeAndContent(parser, SEPARATOR, "[", 1)) {
		Expression *expr = parseExpression(parser);
		if (expr) {
			type->arrayType = createArrayType(expr, type);
			type->type = ARRAY_TYPE_NODE;
		}
		else {
			destroyExpression(expr);
			destroyType(type);
		}
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, "^", 1)) {
		consumeToken(parser); // eat the caret
		type->pointerType = createPointerType(type);
		type->type = POINTER_TYPE_NODE;
		type->pointerType->type = parseType(parser);
	}
	else {
		destroyType(type);
	}

	return type;
}

FieldDecl *parseFieldDecl(Parser *parser) {
	FieldDecl *decl = NULL;

	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (type) {
		IdentifierList *idenList = parseIdentifierList(parser);
		if (idenList) {
			FieldDecl *decl = createFieldDecl(type, mutable);
			decl->idenList = idenList;
		}
		else {
			errorMessage("Failed to parse field items, errored at: `%s`", consumeToken(parser)->content);
			destroyIdentifierList(idenList);
		}
	}
	else {
		errorMessage("Failed to parse type `%s` in Field Declaration", consumeToken(parser)->content);
		destroyType(type);
	}

	return decl;
}

FieldDeclList *parseFieldDeclList(Parser *parser) {
	FieldDeclList *fieldDeclList = createFieldDeclList();

	do {
		FieldDecl *fieldDecl = parseFieldDecl(parser);
		if (fieldDecl) {
			pushBackItem(fieldDeclList->members, fieldDecl);
			if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
				consumeToken(parser);
				if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
					consumeToken(parser);
					break;
				}
			}
			else {
				errorMessage("Expected a semi-colon in field list, but found `%s`", consumeToken(parser)->content);
			}
		}
		else {
			errorMessage("Expected a field declaration, but found `%s`", consumeToken(parser)->content);
			destroyFieldDecl(fieldDecl);
			return NULL;
		}
	}
	while (true);

	return fieldDeclList;
}

StructDecl *parseStructDecl(Parser *parser) {
	StructDecl *structDecl = NULL;

	if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		consumeToken(parser);

		if (checkTokenType(parser, IDENTIFIER, 0)) {
			structDecl = createStructDecl(consumeToken(parser)->content);
			structDecl->fields = parseFieldDeclList(parser);
		}
		else {
			errorMessage("Structure expecting a name but found `%s`", consumeToken(parser)->content);
		}
	}
	else {
		errorMessage("Failed to parse structure errored at: `%s`", consumeToken(parser)->content);
	}

	return structDecl;
}

ParameterSection *parseParameterSection(Parser *parser) {
	ParameterSection *param = NULL;

	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (type) {
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			Token *tok = consumeToken(parser);
			param = createParameterSection(type, mutable);
			param->name = tok->content;
		}
		else {
			errorMessage("Expected an identifier, but found `%s`", consumeToken(parser)->content);
		}
	}
	else {
		errorMessage("Expecting a type, but found `%s`", consumeToken(parser)->content);
		destroyType(type);
	}

	return param;
}

Parameters *parseParameters(Parser *parser) {
	Parameters *params = NULL;

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		params = createParameters();
		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}
			ParameterSection *param = parseParameterSection(parser);
			if (param) {
				pushBackItem(params->paramList, param);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						warningMessage("Trailing comma in parameter list, skipping for now");
					}
					consumeToken(parser);
				}
			}
			else {
				errorMessage("Expected a parameter but found `%s`", consumeToken(parser)->content);
				destroyParameters(params);
				break;
			}
		}
	}
	else {
		errorMessage("Failed to parse parameter list, errored at `%s`", consumeToken(parser)->content);
	}

	return params;
}

Receiver *parseReceiver(Parser *parser) {
	Receiver *receiver = NULL;

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		bool mutable = false;
		if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(parser);
			mutable = true;
		}

		Type *type = parseType(parser);
		if (type) {
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				Token *iden = consumeToken(parser);

				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser);
					receiver = createReceiver(type, iden->content, mutable);
				}
				else {
					errorMessage("Expected closing parenthesis");
				}
			}
			else {
				errorMessage("Expected an identifier in Function Receiver, errored at `%s`", consumeToken(parser)->content);
			}
		}
		else {
			errorMessage("Failed to parse type in Function Receiver, errored at `%s`", consumeToken(parser)->content);
			destroyType(type);
		}
	}

	return receiver;
}

FunctionSignature *parseFunctionSignature(Parser *parser) {
	FunctionSignature *signature = NULL;

	Receiver *receiver = NULL;
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		receiver = parseReceiver(parser);
	}

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *functionName = consumeToken(parser);
		Parameters *params = parseParameters(parser);

		if (params && checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);

			bool mutable = false;
			if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
				consumeToken(parser);
				mutable = true;
			}

			Type *type = parseType(parser);
			if (type) {
				signature = createFunctionSignature(functionName->content, params, mutable, type);
			}
			else {
				destroyType(type);
			}
		}
		else {
			destroyParameters(params);
		}
	}

	return signature;
}

ElseStat *parseElseStat(Parser *parser) {
	ElseStat *stat = NULL;
	if (stat) {
		Block *body = parseBlock(parser);
		if (body) {
			stat = createElseStat();
			stat->body = body;
		}
		else {
			errorMessage("Expected a block in else statement, but found `%s`", consumeToken(parser)->content);
			destroyBlock(body);
		}
	}
	return stat;
}

IfStat *parseIfStat(Parser *parser) {
	IfStat *ifStmt = NULL;

	if (checkTokenTypeAndContent(parser, IDENTIFIER, IF_KEYWORD, 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (expr) {
			Block *block = parseBlock(parser);
			if (block) {
				ElseStat *elseStmt = NULL;
				if (checkTokenTypeAndContent(parser, IDENTIFIER, ELSE_KEYWORD, 0)) {
					elseStmt = parseElseStat(parser);
					if (!elseStmt) {
						errorMessage("Failed to parse else statement");
						return NULL;
					}
				}

				ifStmt = createIfStat();
				ifStmt->body = block;
				ifStmt->elseStmt = elseStmt;
				ifStmt->expr = expr;
			}
			else {
				errorMessage("Expected block after condition in if statement, found `%s`", consumeToken(parser)->content);
				destroyExpression(expr);
				destroyBlock(block);
				return NULL;
			}
		}
		else {
			errorMessage("Expected condition in if statement, found `%s`", consumeToken(parser)->content);
			destroyExpression(expr);
		}
	}

	return ifStmt;
}

ForStat *parseForStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FOR_KEYWORD, 0)) {
		consumeToken(parser);

		Type *type = parseType(parser);
		if (!type) {
			errorMessage("Expected type in for loop signature, found `%s`", consumeToken(parser)->content);
			return NULL;
		}

		Token *index = consumeToken(parser);
		if (index->type != IDENTIFIER) {
			errorMessage("Expected index in for loop signature, found `%s`", index->content);
			destroyType(type);
			return NULL;
		}

		if (checkTokenTypeAndContent(parser, IDENTIFIER, ":", 0)) {
			consumeToken(parser);
		}
		else {
			errorMessage("Expected colon in for loop signature, found `%`", consumeToken(parser)->content);
			destroyType(type);
			return NULL;
		}

		if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
			consumeToken(parser);

			ForStat *forStmt = createForStat(type, index->content);
			int iterations = 0;

			while (true) {
				PrimaryExpr *prim = parsePrimaryExpr(parser);
				if (!prim) {
					errorMessage("Expected primary expression after for loop");
					destroyForStat(forStmt);
					return NULL;
				}
				pushBackItem(forStmt->expr, prim);

				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					if (iterations < 2 || iterations > 3) {
						errorMessage("Too many/few parameters in for loop signature, was given %d arguments, expected 2 or 3", iterations);
						destroyForStat(forStmt);
						return NULL;
					}
					consumeToken(parser);
					break;
				}
				iterations++;
			}

			Block *body = parseBlock(parser);
			if (!body) {
				errorMessage("Expected block after for-loop signature");
				destroyForStat(forStmt);
				return NULL;
			}

			forStmt->body = body;
			return forStmt;
		}
	}

	errorMessage("Failed to parse for loop");
	return NULL;
}

MatchClause *parseMatchClause(Parser *parser) {
	Expression *expr = parseExpression(parser);
	if (!expr) {
		errorMessage("Expected an expression in match clause, found `%s`", consumeToken(parser)->content);
		return NULL;
	}

	if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
		consumeToken(parser);

		Block *block = parseBlock(parser);
		if (!block) {
			errorMessage("Failed to parse block for match clause, errored at `%s`", consumeToken(parser)->content);
			return NULL;
		}

		MatchClause *clause = createMatchClause();
		clause->expr = expr;
		clause->body = block;
		return clause;
	}

	errorMessage("Failed to parse match clause");
	return NULL;
}

MatchStat *parseMatchStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MATCH_KEYWORD, 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (!expr) {
			errorMessage("Failed to parse expression in match statement, errored at `%s`", consumeToken(parser)->content);
			return NULL;
		}

		if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
			consumeToken(parser);

			MatchStat *stmt = createMatchStat(expr);

			int iterations = 0;
			while (true) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
					if (iterations) {
						errorMessage("Match must have at least one match clause, found zero!");
						return NULL;
					}
					consumeToken(parser);
					break;
				}
				iterations++;

				MatchClause *clause = parseMatchClause(parser);
				if (!clause) {
					errorMessage("Failed to parse match clause");
					destroyMatchStat(stmt);
					destroyMatchClause(clause);
					return NULL;
				}
				pushBackItem(stmt->clauses, clause);

				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 1)) {
						warningMessage("Trailing comma in match statement, skipping for now...");
					}
					consumeToken(parser);
				}
			}

			return stmt;
		}
	}

	return NULL;
}

ContinueStat *parseContinueStat(Parser *parser) {
	if (!matchTokenTypeAndContent(parser, IDENTIFIER, CONTINUE_KEYWORD, 0)) {
		errorMessage("Expected continue statement");
		return NULL;
	}
	return createContinueStat();
}

BreakStat *parseBreakStat(Parser *parser) {
	if (!matchTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		errorMessage("Expected break statement");
		return NULL;
	}
	return createBreakStat();
}

ReturnStat *parseReturnStat(Parser *parser) {
	if (!matchTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		errorMessage("Expected return statement");
		return NULL;
	}
	Expression *expr = parseExpression(parser);
	// doesnt matter if there is no expression
	return createReturnStat(expr);
}

LeaveStat *parseLeaveStat(Parser *parser) {
	LeaveStat *leaveStat = createLeaveStat();

	if (checkTokenTypeAndContent(parser, IDENTIFIER, CONTINUE_KEYWORD, 0)) {
		leaveStat->conStmt = parseContinueStat(parser);
		leaveStat->type = CONTINUE_STMT;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		leaveStat->retStmt = parseReturnStat(parser);
		leaveStat->type = RETURN_STMT;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		leaveStat->breakStmt = parseBreakStat(parser);
		leaveStat->type = BREAK_STMT;
	}

	return leaveStat;
}

IncDecStat *parseIncDecStat(Parser *parser) {
	Expression *expr = parseExpression(parser);

	IncOrDec type = IOD_ERRORED;

	if (checkTokenTypeAndContent(parser, OPERATOR, "+", 0) && checkTokenTypeAndContent(parser, OPERATOR, "+", 1)) {
		consumeToken(parser);
		consumeToken(parser);
		type = IOD_INCREMENT;
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, "-", 0) && checkTokenTypeAndContent(parser, OPERATOR, "-", 1)) {
		consumeToken(parser);
		consumeToken(parser);
		type = IOD_DECREMENT;
	}
	else {
		errorMessage("Unknown postfix operator `%s%s`", consumeToken(parser)->content, consumeToken(parser)->content);
		destroyExpression(expr);
		return NULL;
	}

	return createIncDecStat(expr, type);
}

Assignment *parseAssignment(Parser *parser) {
	PrimaryExpr *primaryExpr = parsePrimaryExpr(parser);
	if (!primaryExpr) {
		errorMessage("Expected primary expression");
		return NULL;
	}

	if (!matchTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
		consumeToken(parser);
	}
	else {
		errorMessage("Expected assignment operator");
		destroyPrimaryExpr(primaryExpr);
		return NULL;
	}

	Expression *expr = parseExpression(parser);
	if (!expr) {
		errorMessage("Expected an expression to assign to");
		destroyPrimaryExpr(primaryExpr);
		return NULL;
	}

	return createAssignment(primaryExpr, expr);
}

StructuredStatement *parseStructuredStatement(Parser *parser) {
	StructuredStatement *stmt = createStructuredStatement();

	if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)
			|| checkTokenTypeAndContent(parser, OPERATOR, SINGLE_STATEMENT_OPERATOR, 0)) {
		stmt->block = parseBlock(parser);
		stmt->type = BLOCK_NODE;
		return stmt;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, IF_KEYWORD, 0)) {
		stmt->ifStmt = parseIfStat(parser);
		stmt->type = IF_STAT_NODE;
		return stmt;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, MATCH_KEYWORD, 0)) {
		stmt->matchStmt = parseMatchStat(parser);
		stmt->type = MATCH_STAT_NODE;
		return stmt;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, FOR_KEYWORD, 0)) {
		stmt->forStmt = parseForStat(parser);
		stmt->type = FOR_STAT_NODE;
		return stmt;
	}

	destroyStructuredStatement(stmt);
	return NULL;
}

UnstructuredStatement *parseUnstructuredStatement(Parser *parser) {
	UnstructuredStatement *stmt = createUnstructuredStatement();

	if (checkTokenTypeAndContent(parser, IDENTIFIER, FUNCTION_KEYWORD, 0)
			|| checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)
			|| (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)
			|| checkTokenType(parser, IDENTIFIER, 0)
			|| checkTokenTypeAndContent(parser, SEPARATOR, "[", 1)
			|| checkTokenTypeAndContent(parser, OPERATOR, "^", 1))) {
		stmt->decl = parseDeclaration(parser);
		stmt->type = DECLARATION_NODE;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, CONTINUE_KEYWORD, 0)
			|| checkTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD, 0)
			||checkTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		stmt->leave = parseLeaveStat(parser);
		stmt->type = LEAVE_STAT_NODE;
	}
	else if ((checkTokenTypeAndContent(parser, OPERATOR, "+", 0) && checkTokenTypeAndContent(parser, OPERATOR, "+", 1))
			|| (checkTokenTypeAndContent(parser, OPERATOR, "-", 0) && checkTokenTypeAndContent(parser, OPERATOR, "-", 1))) {
		stmt->incDec = parseIncDecStat(parser);
		stmt->type = INC_DEC_STAT_NODE;
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, "=", 1)) {
		stmt->assignment = parseAssignment(parser);
		stmt->type = ASSIGNMENT_NODE;
	}
	else {
		destroyUnstructuredStatement(stmt);
	}

	return stmt;
}

Statement *parseStatement(Parser *parser) {
	Statement *stmt = createStatement();

	stmt->unstructured = parseUnstructuredStatement(parser);
	if (stmt->unstructured) {
		stmt->type = UNSTRUCTURED_STMT;
		return stmt;
	}

	stmt->structured = parseStructuredStatement(parser);
	if (stmt->structured) {
		stmt->type = STRUCTURED_STMT;
		return stmt;
	}

	errorMessage("O shit stmt failed!\n");
	destroyStatement(stmt);
	return NULL;
}

Block *parseBlock(Parser *parser) {
	StatementList *stmtList = createStatementList();

	if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
		consumeToken(parser);

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
				consumeToken(parser);
				break;
			}

			Statement *stmt = parseStatement(parser);
			if (!stmt) {
				errorMessage("Failed to parse statement");
				destroyStatementList(stmtList);
				return NULL;
			}
			pushBackItem(stmtList->stmts, stmt);

			if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
				consumeToken(parser);
			}
			else {
				errorMessage("Expected a semi-colon at the end of statement, but found `%s`", consumeToken(parser)->content);
				destroyStatementList(stmtList);
				return NULL;
			}
		}

		Block *block = createBlock();
		block->stmtList = stmtList;
		block->type = MULTI_STATEMENT_BLOCK;
		return block;
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, SINGLE_STATEMENT_OPERATOR, 0)) {
		consumeToken(parser);

		Statement *stmt = parseStatement(parser);
		if (!stmt) {
			errorMessage("Failed to parse statement in a single-line block");
			destroyStatementList(stmtList);
			return NULL;
		}
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
			pushBackItem(stmtList->stmts, stmt);
			Block *block = createBlock();
			block->stmtList = stmtList;
			block->type = MULTI_STATEMENT_BLOCK;
			return block;
		}
		else {
			errorMessage("Expected a semi-colon at the end of single-block function, but found `%s`", consumeToken(parser)->content);
			destroyStatementList(stmtList);
			return NULL;
		}
	}

	errorMessage("Failed to parse block");
	return NULL;
}

FunctionDecl *parseFunctionDecl(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FUNCTION_KEYWORD, 0)) {
		consumeToken(parser);

		FunctionSignature *funcSignature = parseFunctionSignature(parser);
		if (!funcSignature) {
			errorMessage("Expected function signature when parsing function declaration");
			return NULL;
		}

		FunctionDecl *funcDecl = createFunctionDecl();
		funcDecl->signature = funcSignature;
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
			return funcDecl;
		}

		Block *block = parseBlock(parser);
		if (!block) {
			errorMessage("Function expected block after function signature but found `%s`", consumeToken(parser)->content);
			destroyFunctionDecl(funcDecl);
			return NULL;
		}

		funcDecl->body = block;
		return funcDecl;
	}

	errorMessage("Failed to parse function declaration");
	return NULL;
}

VariableDecl *parseVariableDecl(Parser *parser) {
	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (!type) {
		errorMessage("Variable Declaration expected a type, but found `%s`", consumeToken(parser)->content);
		return NULL;
	}

	Token *iden = consumeToken(parser);
	if (iden->type != IDENTIFIER) {
		errorMessage("Expected identifier after type in variable declaration but found `%s`", iden->content);
		destroyType(type);
		return NULL;
	}

	Expression *expr = parseExpression(parser);
	if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
		consumeToken(parser);

		if (!expr) {
			errorMessage("Expected expression after assignment operator");
			destroyType(type);
			return NULL;
		}
	}

	if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
		consumeToken(parser);
	}
	else {
		errorMessage("Expected semi-colon at the end of variable declaration");
		destroyType(type);
		destroyExpression(expr);
		return NULL;
	}

	return createVariableDecl(type, iden->content, mutable, expr);
}

Declaration *parseDeclaration(Parser *parser) {
	Declaration *decl = createDeclaration();
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FUNCTION_KEYWORD, 0)) {
		decl->funcDecl = parseFunctionDecl(parser);
		decl->declType = FUNC_DECL;
		if (!decl->funcDecl) {
			errorMessage("Failed to parse function");
			destroyDeclaration(decl);
			return NULL;
		}
		return decl;
	}
	else if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		decl->structDecl = parseStructDecl(parser);
		decl->declType = STRUCT_DECL;
		if (!decl->structDecl) {
			errorMessage("Failed to parse struct declaration");
			destroyDeclaration(decl);
			return NULL;
		}
		return decl;
	}
	else {
		// loads of checks just in case...
		if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)
				|| checkTokenType(parser, IDENTIFIER, 0)
				|| checkTokenTypeAndContent(parser, SEPARATOR, "[", 1)
				|| checkTokenTypeAndContent(parser, OPERATOR, "^", 1)) {
			decl->varDecl = parseVariableDecl(parser);
			decl->declType = VAR_DECL;
			if (!decl->varDecl) {
				errorMessage("Failed to parse variable declaration");
				destroyDeclaration(decl);
				return NULL;
			}
			return decl;
		}

		errorMessage("Unknown declaration");
		destroyDeclaration(decl);
		return NULL;
	}
}

Call *parseCall(Parser *parser) {
	Expression *callee = parseExpression(parser);
	if (!callee) {
		errorMessage("Expected an expression in function callee");
		destroyExpression(callee);
		return NULL;
	}

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		Call *call = createCall(callee);

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			Expression *expr = parseExpression(parser);
			if (!expr) {
				errorMessage("Expected expression in function call");
				destroyExpression(expr);
				destroyExpression(callee);
				destroyCall(call);
				return NULL;
			}
			pushBackItem(call->arguments, expr);

			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					warningMessage("Trailing comma in function call, skipping for now...");
				}
				consumeToken(parser);
			}
			else {
				errorMessage("Expected a comma after expression in function call argument list");
				destroyCall(call);
				return NULL;
			}
		}

		return call;
	}
	else {
		errorMessage("Expected an opening parenthesis in function call");
		destroyExpression(callee);
		return NULL;
	}

	errorMessage("Failed to parse function call");
	destroyExpression(callee);
	return NULL;
}

PrimaryExpr *parsePrimaryExpr(Parser *parser) {
	PrimaryExpr *primaryExpr = createPrimaryExpr();

	// identifier
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		if (!primaryExpr) {
			errorMessage("Faield to parse identifier in primary expression");
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}
		primaryExpr->identifier = consumeToken(parser)->content;
		primaryExpr->type = IDENTIFIER_EXPR;
		return primaryExpr;
	}
	// literal, i.e string, number, character
	else if (isLiteral(parser, 0)) {
		Literal *lit = parseLiteral(parser);
		if (!lit) {
			errorMessage("Failed to parse literal in primary expression");
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}
		primaryExpr->literal = lit;
		primaryExpr->type = LITERAL_EXPR;
		return primaryExpr;
	}
	// paren expr
	else if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);
		Expression *expr = parseExpression(parser);
		if (expr) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				primaryExpr->parenExpr = expr;
				primaryExpr->type = PAREN_EXPR;
				return primaryExpr;
			}
			else {
				errorMessage("Expected a closing parenthesis in expression");
			}
		}
		errorMessage("Failed to parse expression");
	}
	// function call
	else if (checkTokenTypeAndContent(parser, OPERATOR, "(", 1)) {
		Call *call = parseCall(parser);
		if (!call) {
			errorMessage("Failed to parse function call");
			return NULL;
		}
		primaryExpr->funcCall = call;
		primaryExpr->type = FUNC_CALL_EXPR;
		return primaryExpr;
	}
	// array
	else if (checkTokenTypeAndContent(parser, OPERATOR, "[", 1)) {
		Expression *arrName = parseExpression(parser);
		if (!arrName) {
			errorMessage("Expected an expression before array sub-access");
			destroyExpression(arrName);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		if (!matchTokenTypeAndContent(parser, OPERATOR, "[", 0)) {
			errorMessage("Expected an opening square bracket");
			destroyExpression(arrName);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		Expression *length = parseExpression(parser);
		if (!length) {
			errorMessage("Expected a length expression in array sub-access");
			destroyExpression(arrName);
			destroyExpression(length);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		primaryExpr->arraySlice->lhand = arrName;
		primaryExpr->arraySlice->start = length;

		Expression *end = NULL;
		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);

			end = parseExpression(parser);
			if (!end) {
				errorMessage("Expected a cut off point after colon operator in array sub-access");
				destroyExpression(arrName);
				destroyExpression(length);
				destroyExpression(end);
				destroyPrimaryExpr(primaryExpr);
				return NULL;
			}
		}

		primaryExpr->arraySlice->end = end;
		primaryExpr->type = ARRAY_SLICE_EXPR;

		if (!matchTokenTypeAndContent(parser, OPERATOR, "]", 0)) {
			errorMessage("Expected a clsoing square bracket after array sub-access");
			destroyExpression(arrName);
			destroyExpression(length);
			destroyExpression(end);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		return primaryExpr;
	}
	// member access
	else if (checkTokenTypeAndContent(parser, OPERATOR, ".", 1)) {
		Expression *lhand = parseExpression(parser);
		if (!lhand) {
			errorMessage("Expected an expression before array sub-access");
			destroyExpression(lhand);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		if (!matchTokenTypeAndContent(parser, OPERATOR, ".", 0)) {
			errorMessage("Expected a period after expression");
			destroyExpression(lhand);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		if (checkTokenTypeAndContent(parser, OPERATOR, ".", 0)) {
			primaryExpr->memberAccess->value = consumeToken(parser)->content;
		}
		else {
			errorMessage("Expected a period after expression");
			destroyExpression(lhand);
			destroyPrimaryExpr(primaryExpr);
			return NULL;
		}

		primaryExpr->memberAccess->expr = lhand;
		primaryExpr->type = ARRAY_SLICE_EXPR;
		return primaryExpr;
	}

	errorMessage("Failed to parse primary expression");
	destroyPrimaryExpr(primaryExpr);
	return NULL;
}

UnaryExpr *parseUnaryExpr(Parser *parser) {
	Token *tok = consumeToken(parser);
	if (isUnaryOp(tok->content)) {
		Expression *expr = parseExpression(parser);
		if (expr) {
			return createUnaryExpr(tok->content[0], expr);
		}
		destroyExpression(expr);
	}
	return NULL;
}

BinaryExpr *parseBinaryExpr(Parser *parser) {
	Expression *lhand = parseExpression(parser);
	if (lhand) {
		Token *tok = consumeToken(parser);
		if (isBinaryOp(tok->content)) {
			Expression *rhand = parseExpression(parser);
			if (rhand) {
				return createBinaryExpr(lhand, tok->content, rhand);
			}
			destroyExpression(rhand);
		}
	}
	errorMessage("lhand in binary expr failed");
	destroyExpression(lhand);
	return NULL;
}

Expression *parseExpression(Parser *parser) {
	Expression *expr = createExpression();
	expr->binary = parseBinaryExpr(parser);
	if (expr->binary) {
		return expr;
	}

	expr->unary = parseUnaryExpr(parser);
	if (expr->unary) {
		return expr;
	}

	expr->primary = parsePrimaryExpr(parser);
	if (expr->primary) {
		return expr;
	}

	return NULL;
}

/** UTILITIY */

LiteralType getLiteralType(Token *tok) {
	switch (tok->type) {
	case CHARACTER: return LITERAL_CHARACTER;
	case NUMBER: return LITERAL_NUMBER;
	case STRING: return LITERAL_STRING;
	default:
		errorMessage("Unknown literal `%s`", tok->content);
		return LITERAL_ERRORED;
	}
}

Token *consumeToken(Parser *parser) {
	Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex++);
	printf("consumed token: %s, current token is %s\n", tok->content, peekAtTokenStream(parser, 0)->content);
	if (tok->type == END_OF_FILE) parser->parsing = false;
	return tok;
}

bool checkTokenType(Parser *parser, int type, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type;
}

bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type && !strcmp(peekAtTokenStream(parser, 0)->content, content);
}

bool matchTokenType(Parser *parser, int type, int ahead) {
	if (checkTokenType(parser, type, ahead)) {
		consumeToken(parser);
		return true;
	}
	return false;
}

bool matchTokenTypeAndContent(Parser *parser, int type, char *content, int ahead) {
	if (checkTokenTypeAndContent(parser, type, content, ahead)) {
		consumeToken(parser);
		return true;
	}
	return false;
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex + ahead);
}

bool isLiteral(Parser *parser, int ahead) {
	Token *tok = peekAtTokenStream(parser, ahead);
	return tok->type == STRING || tok->type == NUMBER || tok->type == CHARACTER;
}

/** DRIVER */

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
	while (!checkTokenType(parser, END_OF_FILE, parser->tokenIndex)) {
		Statement *stmt = parseStatement(parser);
		if (stmt) {
			pushBackItem(parser->parseTree, stmt);
		}
	}
}
