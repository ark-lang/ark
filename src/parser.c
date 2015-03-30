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
	verboseModeMessage("Destroyed parser");
}

/** PARSING STUFF */

Literal *parseLiteral(Parser *parser) {
	int type = getLiteralType(peekAtTokenStream(parser, 0));
	if (type != ERRORNEOUS) {
		return createLiteral(consumeToken(parser)->content, type);
	}
	return false;
}

IdentifierList *parseIdentifierList(Parser *parser) {
	return false;
}

Type *parseType(Parser *parser) {
	TypeLit *typeLit = parseTypeLit(parser);
	if (typeLit) {
		Type *type = createType();
		type->typeLit = typeLit;
		type->type = TYPE_LIT_NODE;
		return type;
	}

	TypeName *typeName = parseTypeName(parser);
	if (typeName) {
		Type *type = createType();
		type->typeName = typeName;
		type->type = TYPE_NAME_NODE;
		return type;
	}

	return false;
}

FieldDecl *parseFieldDecl(Parser *parser) {
	return false;
}

FieldDeclList *parseFieldDeclList(Parser *parser) {
	return false;
}

StructDecl *parseStructDecl(Parser *parser) {
	return false;
}

ParameterSection *parseParameterSection(Parser *parser) {
	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (type) {
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			char *name = consumeToken(parser)->content;
			ParameterSection *paramSec = createParameterSection(type, mutable);
			paramSec->name = name;
			return paramSec;
		}
	}
	return false;
}

Parameters *parseParameters(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);

		Parameters *params = createParameters();

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
				consumeToken(parser);
				break;
			}

			ParameterSection *paramSection = parseParameterSection(parser);
			if (paramSection) {
				pushBackItem(params->paramList, paramSection);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
						errorMessage("trailing comma");
					}
					consumeToken(parser);
				}
			}
			else {
				break;
			}
		}

		return params;
	}
	return false;
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
		if (type) {
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				char *name = consumeToken(parser)->content;
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser);
				}

				return createReceiver(type, name, mutable);
			}
		}
	}
	return false;
}

FunctionSignature *parseFunctionSignature(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FUNCTION_KEYWORD, 0)) {
		consumeToken(parser);

		Receiver *receiver = parseReceiver(parser);
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			char *functionName = consumeToken(parser)->content;

			if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
				Parameters *params = parseParameters(parser);
				if (params) {
					if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
						consumeToken(parser);

						bool mutable = false;
						if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
							consumeToken(parser);
							mutable = true;
						}

						Type *type = parseType(parser);
						if (type) {
							FunctionSignature *sign = createFunctionSignature(functionName, params, mutable, type);
							sign->receiver = receiver;
							return sign;
						}
					}
				}
			}
		}
	}
	return false;
}

ElseStat *parseElseStat(Parser *parser) {
	return false;
}

IfStat *parseIfStat(Parser *parser) {
	return false;
}

ForStat *parseForStat(Parser *parser) {
	return false;
}

MatchClause *parseMatchClause(Parser *parser) {
	return false;
}

MatchStat *parseMatchStat(Parser *parser) {
	return false;
}

ContinueStat *parseContinueStat(Parser *parser) {
	return false;
}

BreakStat *parseBreakStat(Parser *parser) {
	return false;
}

ReturnStat *parseReturnStat(Parser *parser) {
	return false;
}

LeaveStat *parseLeaveStat(Parser *parser) {
	return false;
}

IncDecStat *parseIncDecStat(Parser *parser) {
	return false;
}

Assignment *parseAssignment(Parser *parser) {
	return false;
}

StructuredStatement *parseStructuredStatement(Parser *parser) {
	return false;
}

UnstructuredStatement *parseUnstructuredStatement(Parser *parser) {
	Declaration *decl = parseDeclaration(parser);
	if (decl) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->decl = decl;
		stmt->type = UNSTRUCTURED_STATEMENT_NODE;
		return stmt;
	}
	return false;
}

Statement *parseStatement(Parser *parser) {
	StructuredStatement *strucStmt = parseStructuredStatement(parser);
	if (strucStmt) {
		Statement *stmt = createStatement();
		stmt->structured = strucStmt;
		stmt->type = STRUCTURED_STATEMENT_NODE;
		return stmt;
	}

	UnstructuredStatement *unstrucStmt = parseUnstructuredStatement(parser);
	if (unstrucStmt) {
		Statement *stmt = createStatement();
		stmt->unstructured = unstrucStmt;
		stmt->type = UNSTRUCTURED_STATEMENT_NODE;
		return stmt;
	}

	return false;
}

Block *parseBlock(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
		consumeToken(parser);

		Block *block = createBlock();
		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
				consumeToken(parser);
				break;
			}

			Statement *stat = parseStatement(parser);
			if (stat) {
				pushBackItem(block->stmtList->stmts, stat);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
						errorMessage("Trailing comma");
					}
					consumeToken(parser);
				}
			}
		}
		return block;
	}
	return false;
}

FunctionDecl *parseFunctionDecl(Parser *parser) {
	FunctionSignature *signature = parseFunctionSignature(parser);
	if (signature) {
		if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
			Block *block = parseBlock(parser);
			if (block) {
				FunctionDecl *decl = createFunctionDecl();
				decl->signature = signature;
				decl->body = block;
				return decl;
			}
		}
	}

	return false;
}

VariableDecl *parseVariableDecl(Parser *parser) {
	bool mutable = false;
	if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(parser);
		mutable = true;
	}

	Type *type = parseType(parser);
	if (type) {
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			char *var_name = consumeToken(parser)->content;
			Expression *rhand = NULL;
			if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
				consumeToken(parser);
				rhand = parseExpression(parser);
			}
			return createVariableDecl(type, var_name, mutable, rhand);
		}
	}

	return false;
}

Declaration *parseDeclaration(Parser *parser) {
	FunctionDecl *func = parseFunctionDecl(parser);
	if (func) {
		Declaration *decl = createDeclaration();
		decl->funcDecl = func;
		decl->type = FUNCTION_DECL_NODE;
		return decl;
	}

	return false;
}

BaseType *parseBaseType(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		BaseType *baseType = createBaseType();
		baseType->type = createTypeName(consumeToken(parser)->content);
		return baseType;
	}

	return false;
}

TypeName *parseTypeName(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		return createTypeName(consumeToken(parser)->content);
	}
	return false;
}

Expression *parseExpression(Parser *parser) {
	UnaryExpr *unary = parseUnaryExpr(parser);
	if (unary) {
		Expression *expr = createExpression();
		expr->unaryExpr = unary;
		expr->type = UNARY_EXPR_NODE;
		return expr;
	}

	Expression *expr = parseExpression(parser);
	if (expr) {
		if (isBinaryOp(peekAtTokenStream(parser, 0)->content)) {
			char *op = consumeToken(parser)->content;
			UnaryExpr *unary = parseUnaryExpr(parser);
			if (unary) {
				Expression *res = createExpression();
				res->lhand = expr;
				res->binaryOp = op;
				res->rhand = unary;
				res->type = BINARY_EXPR_NODE;
				return res;
			}
		}
	}

	return false;
}

ArrayType *parseArrayType(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "[", 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (!expr) {
			destroyExpression(expr);
		}
		else if (checkTokenTypeAndContent(parser, SEPARATOR, "]", 0)) {
			consumeToken(parser);
			Type *type = parseType(parser);
			if (type) {
				return createArrayType(expr, type);
			}
		}
	}

	return false;
}

PointerType *parsePointerType(Parser *parser) {
	if (checkTokenTypeAndContent(parser, OPERATOR, "^", 0)) {
		consumeToken(parser);
		BaseType *type = parseBaseType(parser);
		if (type) {
			return createPointerType(type);
		}
		destroyBaseType(type);
	}
	return false;
}

TypeLit *parseTypeLit(Parser *parser) {
	PointerType *pntr = parsePointerType(parser);
	if (pntr) {
		TypeLit *lit = createTypeLit();
		lit->pointerType = pntr;
		lit->type = POINTER_TYPE_NODE;
		return lit;
	}

	ArrayType *arr = parseArrayType(parser);
	if (arr) {
		TypeLit *lit = createTypeLit();
		lit->arrayType = arr;
		lit->type = ARRAY_TYPE_NODE;
		return lit;
	}

	return false;
}

UnaryExpr *parseUnaryExpr(Parser *parser) {
	PrimaryExpr *prim = parsePrimaryExpr(parser);
	if (prim) {
		UnaryExpr *res = createUnaryExpr();
		res->prim = prim;
		res->type = PRIMARY_EXPR_NODE;
		return res;
	}

	if (isUnaryOp(peekAtTokenStream(parser, 0)->content)) {
		char *op = consumeToken(parser)->content;
		UnaryExpr *unary = parseUnaryExpr(parser);
		if (unary) {
			UnaryExpr *res = createUnaryExpr();
			res->unary = unary;
			res->unaryOp = op;
			res->type = UNARY_EXPR_NODE;
			return res;
		}
	}

	return false;
}

Call *parseCall(Parser *parser) {
	PrimaryExpr *callee = parsePrimaryExpr(parser);
	if (callee) {
		if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
			consumeToken(parser);

			Call *call = createCall(callee);
			while (true) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser);
					break;
				}

				Expression *expr = parseExpression(parser);
				if (expr) {
					pushBackItem(call->arguments, expr);
					if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
						consumeToken(parser);
					}
				}
				else {
					break;
				}
			}

			return call;
		}
	}

	return false;
}

PrimaryExpr *parsePrimaryExpr(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		char *value = consumeToken(parser)->content;
		PrimaryExpr *res = createPrimaryExpr();
		res->identifier = value;
		res->type = IDENTIFIER_NODE;
		return res;
	}

	Literal *lit = parseLiteral(parser);
	if (lit) {
		PrimaryExpr *res = createPrimaryExpr();
		res->literal = lit;
		res->type = LITERAL_NODE;
		return res;
	}

	if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
		consumeToken(parser);
		Expression *expr = parseExpression(parser);
		if (expr) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 1)) {
				consumeToken(parser);
				PrimaryExpr *res = createPrimaryExpr();
				res->parenExpr = expr;
				res->type = PAREN_EXPR_NODE;
				return res;
			}
		}
	}

	Call *call = parseCall(parser);
	if (call) {
		PrimaryExpr *res = createPrimaryExpr();
		res->funcCall = call;
		res->type = FUNCTION_DECL_NODE;
		return res;
	}

	if (checkTokenTypeAndContent(parser, SEPARATOR, "[", 1)) {
		Expression *lhand = parseExpression(parser);
		if (lhand) {
			consumeToken(parser);

			Expression *start = parseExpression(parser);
			if (start) {
				Expression *end = NULL;
				if (checkTokenTypeAndContent(parser, SEPARATOR, ":", 0)) {
					consumeToken(parser);
					end = parseExpression(parser);
				}

				if (checkTokenTypeAndContent(parser, SEPARATOR, "]", 0)) {
					consumeToken(parser);
					PrimaryExpr *res = createPrimaryExpr();
					res->arraySubExpr = createArraySubExpr(lhand);
					res->arraySubExpr->start = start;
					res->arraySubExpr->end = end;
					res->type = ARRAY_SUB_EXPR_NODE;
					return res;
				}
			}
		}
	}
	else if (checkTokenTypeAndContent(parser, SEPARATOR, ".", 1)) {
		Expression *expr = parseExpression(parser);
		if (expr) {
			consumeToken(parser);
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				char *value = consumeToken(parser)->content;
				PrimaryExpr *res = createPrimaryExpr();
				res->memberAccess = createMemberAccessExpr(expr, value);
				res->type = MEMBER_ACCESS_NODE;
				return res;
			}
		}
	}

	return false;
}

/** UTILITIY */

int getLiteralType(Token *tok) {
	switch (tok->type) {
	case CHARACTER: return LITERAL_CHAR;
	case NUMBER: return LITERAL_NUMBER;
	case STRING: return LITERAL_STRING;
	default:
		errorMessage("Unknown literal `%s`", tok->content);
		return LITERAL_ERRORED;
	}
}

Token *consumeToken(Parser *parser) {
	Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex++);
	verboseModeMessage("consumed token: %s, current token is %s", tok->content, peekAtTokenStream(parser, 0)->content);
	if (tok->type == END_OF_FILE) {
		parser->parsing = false;
	}
	return tok;
}

bool checkTokenType(Parser *parser, int type, int ahead) {
	return peekAtTokenStream(parser, ahead)->type == type;
}

bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead) {
	return peekAtTokenStream(parser, ahead)->type == type && !strcmp(peekAtTokenStream(parser, ahead)->content, content);
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
	if (parser->tokenIndex + ahead > parser->tokenStream->size) {
		errorMessage("Attempting to peek at out of bounds token: %d/%d", ahead, parser->tokenStream->size);
		parser->parsing = false;
		return NULL;
	}
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
		parser->parseTree = createVector(VECTOR_EXPONENTIAL);
		parser->tokenIndex = 0;
		parser->parsing = true;

		parseTokenStream(parser);

		file->ast = parser->parseTree;
	}
}

void parseTokenStream(Parser *parser) {
	while (!checkTokenType(parser, END_OF_FILE, 0)) {
		Statement *stmt = parseStatement(parser);
		if (stmt) {
			pushBackItem(parser->parseTree, stmt);
		}
	}
}
