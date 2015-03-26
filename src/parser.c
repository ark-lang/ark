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
	if (type == LITERAL_ERRORED) return NULL;
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
		return type;
	}
	else if (checkTokenTypeAndContent(parser, SEPARATOR, "[", 1)) {
		Type *type = parseType(parser);
		Expression *expr = NULL; // TODO: EXPRESSION!
		type->arrayType = createArrayType(expr, type);
		return type;
	}
	else if (checkTokenTypeAndContent(parser, OPERATOR, "^", 1)) {
		Type *type = parseType(parser);
		consumeToken(parser); // eat the caret
		type->pointerType = createPointerType(type);
		return type;
	}

	errorMessage("Failed to parse invalid type `%s`", consumeToken(parser)->content);
	destroyType(type);
	return NULL;
}

FieldDecl *parseFieldDecl(Parser *parser) {
	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (!type) {
		errorMessage("Failed to parse type `%s` in Field Declaration", consumeToken(parser)->content);
		return NULL;
	}

	IdentifierList *idenList = parseIdentifierList(parser);
	if (!idenList) {
		errorMessage("Failed to parse field items, errored at: `%s`", consumeToken(parser)->content);
		return NULL;
	}

	FieldDecl *decl = createFieldDecl(type, mutable);
	decl->idenList = idenList;
	return decl;
}

FieldDeclList *parseFieldDeclList(Parser *parser) {
	FieldDeclList *fieldDeclList = createFieldDeclList();

	do {
		FieldDecl *fieldDecl = parseFieldDecl(parser);
		if (!fieldDecl) {
			errorMessage("Expected a field declaration, but found `%s`", consumeToken(parser)->content);
			return NULL;
		}
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
	while (true);

	return fieldDeclList;
}

StructDecl *parseStructDecl(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		consumeToken(parser);

		if (checkTokenType(parser, IDENTIFIER, 0)) {
			StructDecl *structDecl = createStructDecl(consumeToken(parser)->content);
			structDecl->fields = parseFieldDeclList(parser);
			return structDecl;
		}
		else {
			errorMessage("Structure expecting a name but found `%s`", consumeToken(parser)->content);
			return NULL;
		}
	}

	errorMessage("Failed to parse structure errored at: `%s`", consumeToken(parser)->content);
	return NULL;
}

ParameterSection *parseParameterSection(Parser *parser) {
	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (!type) {
		errorMessage("Expecting a type, but found `%s`", consumeToken(parser)->content);
		return NULL;
	}

	ParameterSection *param = createParameterSection(type, mutable);
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		param->name = consumeToken(parser)->content;
		return param;
	}

	errorMessage("Expected an identifier, but found `%s`", consumeToken(parser)->content);
	destroyParameterSection(param);
	return NULL;
}

Parameters *parseParameters(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		Parameters  *params = createParameters();

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				return params;
			}
			ParameterSection *param = parseParameterSection(parser);
			if (!param) {
				errorMessage("Expected a parameter but found `%s`", consumeToken(parser)->content);
				destroyParameterSection(param);
				destroyParameters(params);
				return NULL;
			}
			pushBackItem(params->paramList, param);
			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
					warningMessage("Trailing comma in parameter list, skipping for now");
					consumeToken(parser);
				}
				consumeToken(parser);
				break;
			}
		}

		return params;
	}

	errorMessage("Failed to parse parameter list, errored at `%s`", consumeToken(parser)->content);
	return NULL;
}

Receiver *parseReceiver(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		bool mutable = false;
		if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(parser);
			mutable = true;
		}

		Type *type = parseType(parser);
		if (!type) {
			errorMessage("Failed to parse type in Function Receiver, errored at `%s`", consumeToken(parser)->content);
			return NULL;
		}

		if (checkTokenType(parser, IDENTIFIER, 0)) {
			Token *iden = consumeToken(parser);

			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				return createReceiver(type, iden->content, mutable);
			}
		}
		else {
			errorMessage("Expected an identifier in Function Receiver, errored at `%s`", consumeToken(parser)->content);
		}

		destroyType(type);
		return NULL;
	}

	errorMessage("Failed to parse Function Receiver");
	return NULL;
}

FunctionSignature *parseFunctionSignature(Parser *parser) {
	Receiver *receiver = NULL;
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		 receiver = parseReceiver(parser);
	}

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		Token *functionName = consumeToken(parser);

		Parameters *params = parseParameters(parser);
		if (!params) {
			errorMessage("Expected a parameter list for function signature, but found `%s`", consumeToken(parser)->content);
			destroyReceiver(receiver);
			return NULL;
		}

		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);
		}
		else {
			errorMessage("Function signature expected a colon (`:`), but found `%s`", consumeToken(parser)->content);
			destroyReceiver(receiver);
			return NULL;
		}

		bool mutable = false;
		if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(parser);
			mutable = true;
		}

		Type *type = parseType(parser);
		if (!type) {
			errorMessage("Function signature expected a return type, but found `%s`", consumeToken(parser)->content);
			destroyReceiver(receiver);
			return NULL;
		}

		return createFunctionSignature(functionName->content, params, mutable);
	}

	errorMessage("Failed to parse function signature");
	destroyReceiver(receiver);
	return NULL;
}

Statement *parseStatement(Parser *parser) {
	// TODO:
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
			errorMessage("Expected a semi-colon at the end of statement, but found `%s`", consumeToken(parser)->content);
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

		Block *block = parseBlock(parser); // TODO:
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
	return getVectorItem(parser->tokenStream, parser->tokenIndex++);
}

bool checkTokenType(Parser *parser, int type, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type;
}

bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type && !strcmp(peekAtTokenStream(parser, 0)->content, content);
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex + ahead);
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
	while (parser->parsing) {
		Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex);

		switch (tok->type) {
			case IDENTIFIER: break;
			case END_OF_FILE: parser->parsing = false; break;
		}
	}
}
