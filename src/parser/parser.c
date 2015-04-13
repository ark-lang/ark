#include "parser.h"

const char* BINARY_OPS[] = { ".", "*", "/", "%", "+", "-", ">", "<", ">=", "<=",
		"==", "!=", "&", "|", };

const char* DATA_TYPES[] = { "i64", "i32", "i16", "i8", "u64", "u32", "u16",
		"u8", "f64", "f32", "int", "bool", "char", "void" };

int getTypeFromString(char *type) {
	int i;
	for (i = 0; i < ARR_LEN(DATA_TYPES); i++) {
		if (!strcmp(type, DATA_TYPES[i]))
			return i;
	}
	return UNKNOWN_TYPE;
}

Parser *createParser() {
	Parser *parser = safeMalloc(sizeof(*parser));
	parser->tokenStream = NULL;
	parser->tokenIndex = 0;
	parser->parsing = true;
	parser->failed = false;
	parser->binopPrecedence = hashmap_new();

	hashmap_put(parser->binopPrecedence, ".", createPrecedence(4));

	hashmap_put(parser->binopPrecedence, "*", createPrecedence(5));
	hashmap_put(parser->binopPrecedence, "/", createPrecedence(5));
	hashmap_put(parser->binopPrecedence, "%", createPrecedence(5));

	hashmap_put(parser->binopPrecedence, "+", createPrecedence(6));
	hashmap_put(parser->binopPrecedence, "-", createPrecedence(6));

	hashmap_put(parser->binopPrecedence, ">", createPrecedence(8));
	hashmap_put(parser->binopPrecedence, "<", createPrecedence(8));
	hashmap_put(parser->binopPrecedence, ">=", createPrecedence(8));
	hashmap_put(parser->binopPrecedence, "<=", createPrecedence(8));

	hashmap_put(parser->binopPrecedence, "==", createPrecedence(9));
	hashmap_put(parser->binopPrecedence, "!=", createPrecedence(9));

	hashmap_put(parser->binopPrecedence, "&", createPrecedence(10));

	hashmap_put(parser->binopPrecedence, "|", createPrecedence(11));

	return parser;
}

void destroyParser(Parser *parser) {
	if (parser->scope->stackPointer != -1) {
		errorMessage("Some kind of memory leak occurred?");
		while (parser->scope->stackPointer >= 0) {
			popStack(parser->scope);
		}
	}

	destroyStack(parser->scope);
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

UseStatement *parseUseStatement(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, USE_KEYWORD, 0)) {
		consumeToken(parser);

		if (checkTokenType(parser, STRING, 0)) {
			char *file = consumeToken(parser)->content;

			return createUseStatement(file);
		}
	}
	return false;
}

IdentifierList *parseIdentifierList(Parser *parser) {
	IdentifierList *idenList = createIdentifierList();

	while (true) {
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			pushBackItem(idenList->values, consumeToken(parser)->content);
			if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				consumeToken(parser);
			}
		}
		if (!checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
			break;
		}
	}

	return idenList;
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
	bool mutable = false;

	if (checkTokenType(parser, IDENTIFIER, 0)) {
		char *name = consumeToken(parser)->content;

		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);
		}

		if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(parser);
			mutable = true;
		}

		Type *type = parseType(parser);
		if (type) {
			FieldDecl *decl = createFieldDecl(type, mutable);
			decl->name = name;
			return decl;
		}
	}

	return false;
}

FieldDeclList *parseFieldDeclList(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
		consumeToken(parser);

		FieldDeclList *list = createFieldDeclList();
		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
				consumeToken(parser);
				break;
			}

			FieldDecl *decl = parseFieldDecl(parser);
			if (decl) {
				pushBackItem(list->members, decl);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
					consumeToken(parser);
				}
			}
		}
		return list;
	}
	return false;
}

StructDecl *parseStructDecl(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		consumeToken(parser);

		if (checkTokenType(parser, IDENTIFIER, 0)) {
			char *structName = consumeToken(parser)->content;

			if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
				FieldDeclList *list = parseFieldDeclList(parser);
				if (list) {
					StructDecl *decl = createStructDecl(structName);
					decl->fields = list;
					return decl;
				}
			}
		}
	}
	return false;
}

ParameterSection *parseParameterSection(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, "_", 0)) {
		consumeToken(parser);
		ParameterSection *param = createParameterSection(false, true);
		param->name = "_";
		return param;
	}
	else {
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			char *name = consumeToken(parser)->content;
			
			if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
				consumeToken(parser);
			}
			// else oh shit

			bool mutable = false;
			if (checkTokenTypeAndContent(parser, IDENTIFIER, MUT_KEYWORD, 0)) {
				consumeToken(parser);
				mutable = true;
			}

			Type *type = parseType(parser);
			if (type) {
				ParameterSection *paramSec = createParameterSection(type, mutable);
				paramSec->name = name;
				return paramSec;
			}
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
				// its our variadic thing, dont push it back.
				if (!strcmp(paramSection->name, "_")
					&& paramSection->mutable
					&& !paramSection->type) {
					params->variadic = true;
				}
				else {
					pushBackItem(params->paramList, paramSection);
				}

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

FunctionSignature *parseFunctionSignature(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FUNCTION_KEYWORD, 0)) {
		consumeToken(parser);

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
							return sign;
						}
						// else no type specified
					} 
					else if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)
						|| checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
						// just assume it's void.
						Type *type = createType();
						type->typeName = createTypeName(VOID_KEYWORD);
						type->type = TYPE_NAME_NODE;

						FunctionSignature *sign = createFunctionSignature(functionName, params, false, type);
						return sign;
					} 
					else {
						// TODO: colon missing, or block opener missing?
					}
				}
			}
		}
	}
	return false;
}

ElseStat *parseElseStat(Parser *parser) {
	ALLOY_UNUSED_OBJ(parser);
	return false;
}

IfStat *parseIfStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, IF_KEYWORD, 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (expr) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
				Block *block = parseBlock(parser);
				if (block) {
					IfStat *ifStmt = createIfStat();
					ifStmt->expr = expr;
					ifStmt->body = block;
					return ifStmt;
				}
			}
		}
	}
	return false;
}

ForStat *parseForStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, FOR_KEYWORD, 0)) {
		consumeToken(parser);

		// infinite loop
		if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
			Block *block = parseBlock(parser);
			if (block) {
				ForStat *stmt = createForStat();
				stmt->forType = INFINITE_FOR_LOOP;
				stmt->body = block;
				return stmt;
			} else {
				errorMessage("Expected block in for loop");
			}
		}

		Expression *index = parseExpression(parser);
		if (index) {
			// expr {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
				Block *block = parseBlock(parser);
				if (block) {
					ForStat *stmt = createForStat();
					stmt->forType = WHILE_FOR_LOOP;
					stmt->index = index;
					stmt->body = block;
					return stmt;
				} else {
					errorMessage("Expected block in for loop");
				}
			}
			// expr, expr
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
				consumeToken(parser);

				Expression *step = parseExpression(parser);
				if (step) {
					ForStat *stmt = createForStat();
					stmt->index = index;
					stmt->step = step;

					Block *block = parseBlock(parser);
					stmt->forType = INDEX_FOR_LOOP;
					stmt->body = block;
					return stmt;
				}
			} 
			// no fukin clue m8
			else {
				errorMessage("Unknown symbol in for loop");
			}
		}
	}
	return false;
}

MatchClause *parseMatchClause(Parser *parser) {
	ALLOY_UNUSED_OBJ(parser);
	return false;
}

MatchStat *parseMatchStat(Parser *parser) {
	ALLOY_UNUSED_OBJ(parser);
	return false;
}

ContinueStat *parseContinueStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, CONTINUE_KEYWORD, 0)) {
		consumeToken(parser);
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		return createContinueStat();
	}
	return false;
}

BreakStat *parseBreakStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		consumeToken(parser);
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		return createBreakStat();
	}
	return false;
}

ReturnStat *parseReturnStat(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		return createReturnStat(expr);
	}
	return false;
}

LeaveStat *parseLeaveStat(Parser *parser) {
	ContinueStat *cont = parseContinueStat(parser);
	if (cont) {
		LeaveStat *stat = createLeaveStat();
		stat->conStmt = cont;
		stat->type = CONTINUE_STAT_NODE;
		return stat;
	}

	BreakStat *brk = parseBreakStat(parser);
	if (brk) {
		LeaveStat *stat = createLeaveStat();
		stat->breakStmt = brk;
		stat->type = BREAK_STAT_NODE;
		return stat;
	}

	ReturnStat *ret = parseReturnStat(parser);
	if (ret) {
		LeaveStat *stat = createLeaveStat();
		stat->retStmt = ret;
		stat->type = RETURN_STAT_NODE;
		return stat;
	}

	return false;
}

IncDecStat *parseIncDecStat(Parser *parser) {
	Expression *expr = parseExpression(parser);
	if (expr) {
		if (checkTokenTypeAndContent(parser, OPERATOR, "+", 0)
				&& checkTokenTypeAndContent(parser, OPERATOR, "+", 1)) {
			consumeToken(parser);
			consumeToken(parser);
			return createIncDecStat(expr, 1);
		} else if (checkTokenTypeAndContent(parser, OPERATOR, "-", 0)
				&& checkTokenTypeAndContent(parser, OPERATOR, "-", 1)) {
			consumeToken(parser);
			consumeToken(parser);
			return createIncDecStat(expr, -1);
		}
	}
	return false;
}

MemberAccess *parseMemberAccess(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		char *iden = consumeToken(parser)->content;

		MemberExpr *mem = parseMemberExpr(parser);
		if (mem) {
			MemberAccess *access = createMemberAccess();
			access->iden = iden;
			access->expr = mem;
			return access;
		}
	}
	return false;
}

MemberExpr *parseMemberExpr(Parser *parser) {
	Call *call = parseCall(parser);
	if (call) {
		MemberExpr *expr = createMemberExpr();
		expr->call = call;
		expr->type = FUNCTION_CALL_NODE;
		return expr;
	}

	ArrayType *arr = parseArrayType(parser);
	if (arr) {
		MemberExpr *expr = createMemberExpr();
		expr->array = arr;
		expr->type = ARRAY_TYPE_NODE;
		return expr;
	}

	UnaryExpr *unary = parseUnaryExpr(parser);
	if (unary) {
		MemberExpr *expr = createMemberExpr();
		expr->unary = unary;
		expr->type = UNARY_EXPR_NODE;
		return expr;
	}

	MemberAccess *mem = parseMemberAccess(parser);
	if (mem) {
		MemberExpr *expr = createMemberExpr();
		expr->member = mem;
		expr->type = MEMBER_ACCESS_NODE;
		return expr;
	}

	return false;
}

Vector *parseImplBlock(Parser *parser, char *name, char *as) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "{", 0)) {
		consumeToken(parser);

		Vector *v = createVector(VECTOR_EXPONENTIAL);

		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
				consumeToken(parser);
				break;
			}

			FunctionDecl *func = parseFunctionDecl(parser);
			if (func) {
				func->signature->owner = name;
				func->signature->ownerArg = as;
				pushBackItem(v, func);
			}
		}

		return v;
	}

	return false;
}

Impl *parseImpl(Parser *parser) {
	if (checkTokenTypeAndContent(parser, IDENTIFIER, "impl", 0)) {
		consumeToken(parser);

		char *name = NULL;
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			name = consumeToken(parser)->content;
		}
		else {
			errorMessage("NEEDS A NAME PEASANT");
		}

		char *as = NULL;
		if (checkTokenTypeAndContent(parser, IDENTIFIER, "as", 0) && checkTokenType(parser, IDENTIFIER, 1)) {
			consumeToken(parser);
			as = consumeToken(parser)->content;
		}

		Vector *implBlock = parseImplBlock(parser, name, as);
		if (implBlock) {
			Impl *impl = createImpl(name, as);
			impl->funcs = implBlock;
			return impl;
		}

		errorMessage("shite");
	}

	return false;
}

Assignment *parseAssignment(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0) && checkTokenTypeAndContent(parser, OPERATOR, "=", 1)) {
		char *iden = consumeToken(parser)->content;
		if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
			consumeToken(parser);

			Expression *expr = parseExpression(parser);
			if (expr) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
					consumeToken(parser);
				}
				return createAssignment(iden, expr);
			}
		}
	}

	return false;
}

StructuredStatement *parseStructuredStatement(Parser *parser) {
	Block *block = parseBlock(parser);
	if (block) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->block = block;
		stmt->type = BLOCK_NODE;
		return stmt;
	}

	IfStat *ifs = parseIfStat(parser);
	if (ifs) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->ifStmt = ifs;
		stmt->type = IF_STAT_NODE;
		return stmt;
	}

	ForStat *fer = parseForStat(parser);
	if (fer) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->forStmt = fer;
		stmt->type = FOR_STAT_NODE;
		return stmt;
	}

	return false;
}

UnstructuredStatement *parseUnstructuredStatement(Parser *parser) {
	UseStatement *use = parseUseStatement(parser);
	if (use) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->use = use;
		stmt->type = USE_STATEMENT_NODE;
		return stmt;
	}

	Impl *impl = parseImpl(parser);
	if (impl) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->impl = impl;
		stmt->type = IMPL_NODE;
		return stmt;
	}

	LeaveStat *leave = parseLeaveStat(parser);
	if (leave) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->leave = leave;
		stmt->type = LEAVE_STAT_NODE;
		return stmt;
	}

	Call *call = parseCall(parser);
	if (call) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->call = call;
		stmt->type = FUNCTION_CALL_NODE;
		if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);
		}
		else {
			// FIXME EXPECTED SEMI COLON
		}
		return stmt;
	}

	Assignment *assign = parseAssignment(parser);
	if (assign) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->assignment = assign;
		stmt->type = ASSIGNMENT_NODE;
		return stmt;
	}

	Declaration *decl = parseDeclaration(parser);
	if (decl) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->decl = decl;
		stmt->type = DECLARATION_NODE;
		return stmt;
	}

	IncDecStat *incDec = parseIncDecStat(parser);
	if (incDec) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->incDec = incDec;
		stmt->type = INC_DEC_STAT_NODE;
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
		pushScope(parser);
		consumeToken(parser);

		Block *block = createBlock();
		while (true) {
			if (checkTokenTypeAndContent(parser, SEPARATOR, "}", 0)) {
				popScope(parser, block);
				consumeToken(parser);
				break;
			}

			Statement *stat = parseStatement(parser);
			if (stat) {
				pushBackItem(block->stmtList->stmts, stat);
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
		} else if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
			consumeToken(parser);

			FunctionDecl *decl = createFunctionDecl();
			decl->signature = signature;
			decl->prototype = true;
			return decl;
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

	// int x = expr;

	// x: int = expr;
	if (checkTokenType(parser, IDENTIFIER, 0)) {
		char *var_name = consumeToken(parser)->content;
		Expression *rhand = NULL;

		if (checkTokenTypeAndContent(parser, OPERATOR, ":", 0)) {
			consumeToken(parser);
			
			Type *type = parseType(parser);
			if (!type) {
				errorMessage("NO TYPE!");
			}

			if (checkTokenTypeAndContent(parser, OPERATOR, "=", 0)) {
				consumeToken(parser);

				rhand = parseExpression(parser);
				if (rhand) {
					if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
						consumeToken(parser);
					}

					VariableDecl *decl = createVariableDecl(type, var_name, mutable, rhand);
					decl->assigned = true;
					return decl;
				}
			}
			else if (checkTokenTypeAndContent(parser, SEPARATOR, ";", 0)) {
				consumeToken(parser); // eat the semi colon!

				VariableDecl *decl = createVariableDecl(type, var_name, mutable, rhand);
				decl->assigned = false;
				return decl;
			}
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

	StructDecl *struc = parseStructDecl(parser);
	if (struc) {
		Declaration *decl = createDeclaration();
		decl->structDecl = struc;
		decl->type = STRUCT_DECL_NODE;
		return decl;
	}

	VariableDecl *varDecl = parseVariableDecl(parser);
	if (varDecl) {
		Declaration *decl = createDeclaration();
		decl->varDecl = varDecl;
		decl->type = VARIABLE_DECL_NODE;
		return decl;
	}

	return false;
}

int getTokenPrecedence(Parser *parser) {
	Token *tok = peekAtTokenStream(parser, 0);

	if (!isASCII(tok->content[0]))
		return -1;

	Precedence *prec = NULL;
	if (hashmap_get(parser->binopPrecedence, tok->content,
			(void**) &prec) == MAP_MISSING) {
		verboseModeMessage("Precedence doesnt exist for %s\n", tok->content);
		return -1;
	}

	int tokenPrecedence = prec->prec;
	if (tokenPrecedence <= 0) {
		return -1;
	}
	return tokenPrecedence;
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
	Expression *expr = parsePrimaryExpression(parser);
	if (!expr)
		return false;

	if (isValidBinaryOp(peekAtTokenStream(parser, 0)->content)) {
		return parseBinaryOperator(parser, 0, expr);
	}
	return expr;
}

ArrayType *parseArrayType(Parser *parser) {
	if (checkTokenTypeAndContent(parser, SEPARATOR, "[", 0)) {
		consumeToken(parser);

		Expression *expr = parseExpression(parser);
		if (!expr) {
			destroyExpression(expr);
		} else if (checkTokenTypeAndContent(parser, SEPARATOR, "]", 0)) {
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
	if (isUnaryOp(peekAtTokenStream(parser, 0)->content)) {
		char *op = consumeToken(parser)->content;
		Expression *prim = parsePrimaryExpression(parser);
		if (prim) {
			UnaryExpr *res = createUnaryExpr();
			res->lhand = prim;
			res->unaryOp = op;
			return res;
		}
	}

	return false;
}

Expression *parseBinaryOperator(Parser *parser, int precedence,
		Expression *lhand) {
	for (;;) {
		int tokenPrecedence = getTokenPrecedence(parser);
		if (tokenPrecedence < precedence)
			return lhand;

		Token *tok = peekAtTokenStream(parser, 0);
		if (!isValidBinaryOp(tok->content)) {
			errorMessage("No precedence for %s", tok->content);
			return false;
		}
		char *binaryOp = consumeToken(parser)->content;

		Expression *rhand = parsePrimaryExpression(parser);
		if (!rhand)
			return false;

		int nextPrec = getTokenPrecedence(parser);
		if (tokenPrecedence < nextPrec) {
			rhand = parseBinaryOperator(parser, tokenPrecedence + 1, rhand);
			if (!rhand)
				return false;
		}

		Expression *temp = createExpression();
		temp->binary = createBinaryExpr();
		temp->binary->lhand = lhand;
		temp->binary->binaryOp = binaryOp;
		temp->binary->rhand = rhand;
		temp->exprType = BINARY_EXPR_NODE;
		lhand = temp;
	}

	return false;
}

Expression *parsePrimaryExpression(Parser *parser) {
	if(checkTokenType(parser, IDENTIFIER, 0) && (checkTokenTypeAndContent(parser, SEPARATOR, "(", 1))) {
		Call *call = parseCall(parser);
		if (call) {
			Expression *expr = createExpression();
			expr->call = call;
			expr->exprType = FUNCTION_CALL_NODE;
			return expr;
		}
	}

	if (checkTokenType(parser, IDENTIFIER, 0) 
		&& checkTokenTypeAndContent(parser, SEPARATOR, ".", 1)) {
		Vector *members = createVector(VECTOR_EXPONENTIAL);
		while (true) {
			if (checkTokenType(parser, IDENTIFIER, 0)) {
				char *iden = consumeToken(parser)->content;
				pushBackItem(members, iden);
				if (checkTokenTypeAndContent(parser, SEPARATOR, ".", 0)) {
					consumeToken(parser);
				}
				else {
					break;
				}
			}

			for (int i = 0; i < members->size; i++) {
				printf("%s\n", getVectorItem(members, i));
			}
		}
	}

	UnaryExpr *unary = parseUnaryExpr(parser);
	if (unary) {
		Expression *expr = createExpression();
		expr->unary = unary;
		expr->exprType = UNARY_EXPR_NODE;
		return expr;
	}

	Type *type = parseType(parser);
	if (type) {
		Expression *expr = createExpression();
		expr->type = type;
		expr->exprType = TYPE_NODE;
		return expr;
	}

	Literal *lit = parseLiteral(parser);
	if (lit) {
		Expression *expr = createExpression();
		expr->lit = lit;
		expr->exprType = LITERAL_NODE;
		return expr;
	}

	return false;
}

Call *parseCall(Parser *parser) {
	if (checkTokenType(parser, IDENTIFIER, 0) && 
		(checkTokenTypeAndContent(parser, SEPARATOR, "(", 1) || checkTokenTypeAndContent(parser, SEPARATOR, ".", 1))) {

		Vector *idens = NULL;
		if (checkTokenType(parser, IDENTIFIER, 0)) {
			idens = createVector(VECTOR_LINEAR);
			while(true) {
				if (checkTokenType(parser, IDENTIFIER, 0)) {
					pushBackItem(idens, consumeToken(parser)->content);
				}
				if (checkTokenTypeAndContent(parser, SEPARATOR, ".", 0)) {
					consumeToken(parser);
				}
				if (!checkTokenType(parser, IDENTIFIER, 0) || !checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
					break;
				}
			}
		}

		if (checkTokenTypeAndContent(parser, SEPARATOR, "(", 0)) {
			consumeToken(parser);

			Call *call = createCall(idens);
			while (true) {
				if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 0)) {
					consumeToken(parser);
					break;
				}

				Expression *expr = parseExpression(parser);
				if (expr) {
					pushBackItem(call->arguments, expr);
					if (checkTokenTypeAndContent(parser, SEPARATOR, ",", 0)) {
						if (checkTokenTypeAndContent(parser, SEPARATOR, ")", 1)) {
							errorMessage("Warning, trailing comma in function call. Skipping.\n");
						}
						consumeToken(parser);
					}
				}
			}
			return call;
		}

	}
	return false;
}

/** UTILITY */

void pushScope(Parser *parser) {
	pushToStack(parser->scope, createScope());
}

void pushPointer(Parser *parser, char *name) {
	Scope *scope = getStackItem(parser->scope, parser->scope->stackPointer);
	pushBackItem(scope->pointers, createPointerFree(name));
}

void popScope(Parser *parser, Block *block) {
	Scope *scope = popStack(parser->scope);

	for (int i = 0; i < scope->pointers->size; i++) {
		PointerFree *pntr = getVectorItem(scope->pointers, i);

		UnstructuredStatement *unstructured = createUnstructuredStatement();
		unstructured->pointerFree = pntr;
		unstructured->type = POINTER_FREE_NODE;

		Statement *stmt = createStatement();
		stmt->unstructured = unstructured;
		stmt->type = UNSTRUCTURED_STATEMENT_NODE;

		pushBackItem(block->stmtList->stmts, stmt);
	}
}

int getLiteralType(Token *tok) {
	switch (tok->type) {
	case CHARACTER:
		return LITERAL_CHAR;
	case NUMBER:
		return LITERAL_NUMBER;
	case STRING:
		return LITERAL_STRING;
	default:
		errorMessage("Unknown literal `%s`", tok->content);
		return LITERAL_ERRORED;
	}
}

Token *consumeToken(Parser *parser) {
	Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex++);
	verboseModeMessage("consumed token: %s, current token is %s", tok->content,
			peekAtTokenStream(parser, 0)->content);
	if (tok->type == END_OF_FILE) {
		parser->parsing = false;
	}
	return tok;
}

bool checkTokenType(Parser *parser, int type, int ahead) {
	return peekAtTokenStream(parser, ahead)->type == type;
}

bool checkTokenTypeAndContent(Parser *parser, int type, char *content,
		int ahead) {
	return peekAtTokenStream(parser, ahead)->type == type
			&& !strcmp(peekAtTokenStream(parser, ahead)->content, content);
}

bool matchTokenType(Parser *parser, int type, int ahead) {
	if (checkTokenType(parser, type, ahead)) {
		consumeToken(parser);
		return true;
	}
	return false;
}

bool matchTokenTypeAndContent(Parser *parser, int type, char *content,
		int ahead) {
	if (checkTokenTypeAndContent(parser, type, content, ahead)) {
		consumeToken(parser);
		return true;
	}
	return false;
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	if (parser->tokenIndex + ahead > parser->tokenStream->size) {
		errorMessage("Attempting to peek at out of bounds token: %d/%d", ahead,
				parser->tokenStream->size);
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
		parser->scope = createStack();

		parseTokenStream(parser);

		file->ast = parser->parseTree;
	}
}

bool isValidBinaryOp(char *tok) {
	int size = ARR_LEN(BINARY_OPS);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok, BINARY_OPS[i])) {
			return true;
		}
	}
	return false;
}

void parseTokenStream(Parser *parser) {
	while (!checkTokenType(parser, END_OF_FILE, 0)) {
		Statement *stmt = parseStatement(parser);
		if (stmt) {
			pushBackItem(parser->parseTree, stmt);
		}
	}
}
