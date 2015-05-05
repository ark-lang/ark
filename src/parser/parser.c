#include "parser.h"

#define parserError(...) self->failed = true; \
						 errorMessage(__VA_ARGS__)

#define PRINT_CURR_TOK() printf("current token: %s\n", peekAtTokenStream(self, 0)->content);

const char* BINARY_OPS[] = { ".", "*", "/", "%", "+", "=", "-", ">", "<", ">=", "<=", "==", "!=", "&", "|", };

const char* DATA_TYPES[] = { "i64", "i32", "i16", "i8", "u64", "u32", "u16", "u8", "f64", "f32", "int", "bool", "char", "void" };

int getTypeFromString(char *type) {
	for (int i = 0; i < ARR_LEN(DATA_TYPES); i++) {
		if (!strcmp(type, DATA_TYPES[i]))
			return i;
	}
	return UNKNOWN_TYPE;
}

Parser *createParser() {
	Parser *self = safeMalloc(sizeof(*self));
	self->tokenStream = NULL;
	self->tokenIndex = 0;
	self->parsing = true;
	self->failed = false;
	self->binopPrecedence = hashmap_new();

	hashmap_put(self->binopPrecedence, "++", createPrecedence(3));
	hashmap_put(self->binopPrecedence, "--", createPrecedence(3));
	hashmap_put(self->binopPrecedence, "!", createPrecedence(3));
	hashmap_put(self->binopPrecedence, "~", createPrecedence(3));
	hashmap_put(self->binopPrecedence, "&", createPrecedence(3));

	hashmap_put(self->binopPrecedence, ".", createPrecedence(4));

	hashmap_put(self->binopPrecedence, "*", createPrecedence(5));
	hashmap_put(self->binopPrecedence, "/", createPrecedence(5));
	hashmap_put(self->binopPrecedence, "%", createPrecedence(5));


	hashmap_put(self->binopPrecedence, "+", createPrecedence(6));
	hashmap_put(self->binopPrecedence, "-", createPrecedence(6));

	hashmap_put(self->binopPrecedence, ">", createPrecedence(8));
	hashmap_put(self->binopPrecedence, "<", createPrecedence(8));
	hashmap_put(self->binopPrecedence, ">=", createPrecedence(8));
	hashmap_put(self->binopPrecedence, "<=", createPrecedence(8));

	hashmap_put(self->binopPrecedence, "==", createPrecedence(9));
	hashmap_put(self->binopPrecedence, "!=", createPrecedence(9));

	hashmap_put(self->binopPrecedence, "&", createPrecedence(10));

	hashmap_put(self->binopPrecedence, "|", createPrecedence(11));

	hashmap_put(self->binopPrecedence, "=", createPrecedence(15));

	return self;
}

void destroyParser(Parser *self) {
	free(self);
	verboseModeMessage("Destroyed self");
}

/** PARSING STUFF */

Literal *parseLiteral(Parser *self) {
	int type = getLiteralType(peekAtTokenStream(self, 0));
	if (type != ERRORNEOUS) {
		return createLiteral(consumeToken(self)->content, type);
	}
	return false;
}

UseMacro *parseUseMacro(Parser *self) {
	if (checkTokenTypeAndContent(self, OPERATOR, "!", 0)
		&& checkTokenTypeAndContent(self, IDENTIFIER, USE_KEYWORD, 1)) {
		consumeToken(self);
		consumeToken(self);

		UseMacro *use = createUseMacro();

		if (checkTokenType(self, STRING, 0)) {
			char *fileName = consumeToken(self)->content;
			pushBackItem(use->files, fileName);
			return use;
		}
		else if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
			consumeToken(self);

			while (true) {
				if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
					consumeToken(self);
					break;
				}

				if (checkTokenType(self, STRING, 0)) {
					char *fileName = consumeToken(self)->content;
					if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
						consumeToken(self);
					}
					pushBackItem(use->files, fileName);
				}
			}

			return use;
		}
	}
	return false;
}

LinkerFlagMacro *parseLinkerFlagMacro(Parser *self) {
	if (checkTokenTypeAndContent(self, OPERATOR, "!", 0)
		&& checkTokenTypeAndContent(self, IDENTIFIER, LINKER_FLAG_KEYWORD, 1)) {
		consumeToken(self);
		consumeToken(self);

		if (checkTokenType(self, STRING, 0)) {
			char *flag = consumeToken(self)->content;
			return createLinkerFlagMacro(flag);
		}
	}
	return false;
}

IdentifierList *parseIdentifierList(Parser *self) {
	IdentifierList *idenList = createIdentifierList();

	while (true) {
		if (checkTokenType(self, IDENTIFIER, 0)) {
			pushBackItem(idenList->values, consumeToken(self)->content);
			if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
				consumeToken(self);
			}
		}
		if (!checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
			break;
		}
	}

	return idenList;
}

Type *parseType(Parser *self) {
	TypeLit *typeLit = parseTypeLit(self);
	if (typeLit) {
		Type *type = createType();
		type->typeLit = typeLit;
		type->type = TYPE_LIT_NODE;
		return type;
	}

	TypeName *typeName = parseTypeName(self);
	if (typeName) {
		Type *type = createType();
		type->typeName = typeName;
		type->type = TYPE_NAME_NODE;
		return type;
	}
	
	return false;
}

FieldDecl *parseFieldDecl(Parser *self) {
	bool mutable = false;

	if (checkTokenType(self, IDENTIFIER, 0)) {
		char *name = consumeToken(self)->content;

		if (checkTokenTypeAndContent(self, OPERATOR, ":", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Expected colon in field declaration, found: %s", peekAtTokenStream(self, 0)->content);
		}

		if (checkTokenTypeAndContent(self, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(self);
			mutable = true;
		}

		Type *type = parseType(self);
		if (type) {
			FieldDecl *decl = createFieldDecl(type, mutable);
			decl->name = name;
			return decl;
		}
		else {
			parserError("Expected type in field declaration, found: %s", peekAtTokenStream(self, 0)->content);
		}
	}

	return false;
}

FieldDeclList *parseFieldDeclList(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
		consumeToken(self);

		FieldDeclList *list = createFieldDeclList();
		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
				consumeToken(self);
				break;
			}

			FieldDecl *decl = parseFieldDecl(self);
			if (decl) {
				pushBackItem(list->members, decl);
				if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(self, SEPARATOR, "}", 1)) {
						parserError("Trailing comma in field declaration, found: %s", peekAtTokenStream(self, 0)->content);
					}
					consumeToken(self);
				}
				else if (!checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
					parserError("Expected comma after field declaration, found: %s", peekAtTokenStream(self, 0)->content);
				}
			}
		}
		return list;
	}
	return false;
}

StructDecl *parseStructDecl(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		consumeToken(self);

		if (checkTokenType(self, IDENTIFIER, 0)) {
			char *structName = consumeToken(self)->content;

			if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
				FieldDeclList *list = parseFieldDeclList(self);
				if (list) {
					StructDecl *decl = createStructDecl(structName);
					decl->fields = list;
					return decl;
				}
			}
			else {
				parserError("Expected body for structure `%s`, found: %s", structName, peekAtTokenStream(self, 0)->content);
			}
		}
	}
	return false;
}

ParameterSection *parseParameterSection(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, "_", 0)) {
		consumeToken(self);
		ParameterSection *param = createParameterSection(false, true);
		param->name = "_";
		return param;
	}
	else {
		bool mutable = false;
		if (checkTokenTypeAndContent(self, IDENTIFIER, MUT_KEYWORD, 0)) {
			consumeToken(self);
			mutable = true;
		}

		if (checkTokenType(self, IDENTIFIER, 0)) {
			char *name = consumeToken(self)->content;
			
			if (checkTokenTypeAndContent(self, OPERATOR, ":", 0)) {
				consumeToken(self);
			}
			else {
				parserError("Expected colon after argument identifier, found: %s", peekAtTokenStream(self, 0)->content);
			}

			Type *type = parseType(self);
			if (!type) {
				parserError("Expected Type after colon in argument, found: %s", peekAtTokenStream(self, 0)->content);
			}

			ParameterSection *paramSec = createParameterSection(type, mutable);
			paramSec->name = name;
			return paramSec;
		}
	}
	return false;
}

Parameters *parseParameters(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
		consumeToken(self);

		Parameters *params = createParameters();

		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, ")", 0)) {
				consumeToken(self);
				break;
			}

			ParameterSection *paramSection = parseParameterSection(self);
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

				if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
					if (checkTokenTypeAndContent(self, SEPARATOR, ")", 1)) {
						parserError("Trailing comma in parameter list, found: %s", peekAtTokenStream(self, 0)->content);
					}
					consumeToken(self);
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

FunctionSignature *parseFunctionSignature(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, FUNCTION_KEYWORD, 0)) {
		consumeToken(self);

		if (checkTokenType(self, IDENTIFIER, 0)) {
			char *functionName = consumeToken(self)->content;

			if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
				Parameters *params = parseParameters(self);
				if (params) {
					if (checkTokenTypeAndContent(self, OPERATOR, ":", 0)) {
						consumeToken(self);

						bool mutable = false;
						if (checkTokenTypeAndContent(self, IDENTIFIER, MUT_KEYWORD, 0)) {
							consumeToken(self);
							mutable = true;
						}

						Type *type = parseType(self);
						if (type) {
							FunctionSignature *sign = createFunctionSignature(functionName, params, mutable, type);
							return sign;
						}
						else {
							parserError("Expected function return type in `%s`, found: %s", functionName, peekAtTokenStream(self, 0)->content);
						}
					} 
					else if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)
						|| checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
						// just assume it's void.
						Type *type = createType();
						type->typeName = createTypeName(VOID_KEYWORD);
						type->type = TYPE_NAME_NODE;

						FunctionSignature *sign = createFunctionSignature(functionName, params, false, type);
						return sign;
					} 
					else {
						parserError("Function `%s` is missing a block, return type, or semi-colon, found: %s", functionName, peekAtTokenStream(self, 0)->content);
					}
				}
			}
		}
	}
	return false;
}

EnumItem *parseEnumItem(Parser *self) {
	// a = expr,
	// b,
	
	if (checkTokenType(self, IDENTIFIER, 0)) {
		char *itemName = consumeToken(self)->content;

		EnumItem *item = createEnumItem(itemName);
		if (checkTokenTypeAndContent(self, OPERATOR, "=", 0)) {
			consumeToken(self);

			Expression *expr = parseExpression(self);
			if (expr) {
				item->val = expr;
			}
			else {
				parserError("Expected expression after assignment operator for enumerator item `%s`", item);
			}
		}
		return item;
	}

	return false;
}

EnumDecl *parseEnumDecl(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, ENUM_KEYWORD, 0)) {
		consumeToken(self);

		if (checkTokenType(self, IDENTIFIER, 0)) {
			char *enumName = consumeToken(self)->content;

			if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
				consumeToken(self);

				EnumDecl *enumDecl = createEnumDecl(enumName);
				while (true) {
					if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
						consumeToken(self);
						break;
					}

					EnumItem *item = parseEnumItem(self);
					if (item) {
						pushBackItem(enumDecl->items, item);
						if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
							if (checkTokenTypeAndContent(self, SEPARATOR, "}", 1)) {
								parserError("Trailing comma in enumerator `%s`, found: %s", enumName, peekAtTokenStream(self, 0)->content);
							}
							consumeToken(self);
						}
					}
				}

				return enumDecl;
			}
		}
		else {
			parserError("Enumerator expected a name, found: %s", peekAtTokenStream(self, 0)->content);
		}
	}

	return false;
}

ElseStat *parseElseStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, ELSE_KEYWORD, 0)) {
		consumeToken(self);

		if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
			Block *block = parseBlock(self);
			if (block) {
				ElseStat *elseStat = createElseStat();
				elseStat->body = block;
				return elseStat;
			}
			else {
				parserError("Failed to parse block");
			}
		}
	}
	return false;
}

ElseIfStat *parseElseIfStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, ELSE_KEYWORD, 0)) {
		consumeToken(self);

		if (checkTokenTypeAndContent(self, IDENTIFIER, IF_KEYWORD, 0)) {
			consumeToken(self);

			if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
				Block *block = parseBlock(self);
				if (block) {
					ElseIfStat *elif = createElseIfStat(self);

					return elif;
				}
			}
		}
	}

	return false;
}

IfStat *parseIfStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, IF_KEYWORD, 0)) {
		consumeToken(self);

		Expression *expr = parseExpression(self);
		if (expr) {
			if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
				Block *block = parseBlock(self);
				if (block) {
					IfStat *ifStmt = createIfStat();
					if (checkTokenTypeAndContent(self, IDENTIFIER, ELSE_KEYWORD, 0)
						&& checkTokenTypeAndContent(self, SEPARATOR, "{", 1)) {
						ifStmt->elseStmt = parseElseStat(self);
					}
					else if (checkTokenTypeAndContent(self, IDENTIFIER, ELSE_KEYWORD, 0)
						&& checkTokenTypeAndContent(self, IDENTIFIER, IF_KEYWORD, 1)) {

					}
					ifStmt->expr = expr;
					ifStmt->body = block;
					return ifStmt;
				}
			}
			else {
				parserError("If Statement expected a block, found: %s", peekAtTokenStream(self, 0)->content);
			}
		}
	}
	return false;
}

ForStat *parseForStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, FOR_KEYWORD, 0)) {
		consumeToken(self);

		// infinite loop
		if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
			Block *block = parseBlock(self);
			if (block) {
				ForStat *stmt = createForStat();
				stmt->forType = INFINITE_FOR_LOOP;
				stmt->body = block;
				return stmt;
			} 
			else {
				parserError("For Loop expected a block, found: %s", peekAtTokenStream(self, 0)->content);
			}
		}

		Expression *index = parseExpression(self);
		if (index) {
			// expr {
			if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
				Block *block = parseBlock(self);
				if (block) {
					ForStat *stmt = createForStat();
					stmt->forType = WHILE_FOR_LOOP;
					stmt->index = index;
					stmt->body = block;
					return stmt;
				} 
				else {
					parserError("For Loop expected a block, found: %s", peekAtTokenStream(self, 0)->content);
				}
			}
			// expr, expr
			else if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
				consumeToken(self);

				Expression *step = parseExpression(self);
				if (step) {
					ForStat *stmt = createForStat();
					stmt->index = index;
					stmt->step = step;

					Block *block = parseBlock(self);
					stmt->forType = INDEX_FOR_LOOP;
					stmt->body = block;
					return stmt;
				}
			} 
			else {
				parserError("Invalid signature in for loop, found: %s", peekAtTokenStream(self, 0)->content);
			}
		}
	}

	return false;
}

MatchClause *parseMatchClause(Parser *self) {

    Expression *expr = parseExpression(self);
    if (expr) {
	    MatchClause *clause = createMatchClause();
        if (clause) {
        	Block *block = parseBlock(self);
            if (block) {
                clause->expr = expr;
                clause->body = block;
                return clause;
            }
        }
    }

    return false;
}

MatchStat *parseMatchStat(Parser *self) {
    if (checkTokenTypeAndContent(self, IDENTIFIER, MATCH_KEYWORD, 0)) {
        consumeToken(self);
        
        Expression *expr = parseExpression(self);
        if (expr) {
        	MatchStat *stmt = createMatchStat(expr);
	        if (!stmt) {
	        	verboseModeMessage("Failed to create statement");
	        	return false;
	        }

        	if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
	            consumeToken(self);
	            
	            while (true) {
		        	if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
		            	consumeToken(self);
		            	break;
		            }

		            MatchClause *clause = parseMatchClause(self);
		            if (clause) {
		                pushBackItem(stmt->clauses, clause);
		            }
	            }

	            return stmt;
	        }
        }
    }

    return false;
}

ContinueStat *parseContinueStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, CONTINUE_KEYWORD, 0)) {
		consumeToken(self);
		if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Expected semi-colon at end of continue statement, found: %s", peekAtTokenStream(self, 0)->content);
		}
		return createContinueStat();
	}
	return false;
}

BreakStat *parseBreakStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, BREAK_KEYWORD, 0)) {
		consumeToken(self);
		if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Expected semi-colon at end of break statement, found: %s", peekAtTokenStream(self, 0)->content);
		}
		return createBreakStat();
	}
	return false;
}

ReturnStat *parseReturnStat(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, RETURN_KEYWORD, 0)) {
		consumeToken(self);

		Expression *expr = parseExpression(self);
		if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Expected semi-colon at end of return statement, found: %s", peekAtTokenStream(self, 0)->content);
		}
		return createReturnStat(expr);
	}
	return false;
}

LeaveStat *parseLeaveStat(Parser *self) {
	ContinueStat *cont = parseContinueStat(self);
	if (cont) {
		LeaveStat *stat = createLeaveStat();
		stat->conStmt = cont;
		stat->type = CONTINUE_STAT_NODE;
		return stat;
	}

	BreakStat *brk = parseBreakStat(self);
	if (brk) {
		LeaveStat *stat = createLeaveStat();
		stat->breakStmt = brk;
		stat->type = BREAK_STAT_NODE;
		return stat;
	}

	ReturnStat *ret = parseReturnStat(self);
	if (ret) {
		LeaveStat *stat = createLeaveStat();
		stat->retStmt = ret;
		stat->type = RETURN_STAT_NODE;
		return stat;
	}

	return false;
}

// TODO FIXME
IncDecStat *parseIncDecStat(Parser *self) {
	Expression *expr = parseExpression(self);
	if (expr) {
		if (checkTokenTypeAndContent(self, OPERATOR, "+", 0)
				&& checkTokenTypeAndContent(self, OPERATOR, "+", 1)) {
			consumeToken(self);
			consumeToken(self);
			return createIncDecStat(expr, 1);
		} else if (checkTokenTypeAndContent(self, OPERATOR, "-", 0)
				&& checkTokenTypeAndContent(self, OPERATOR, "-", 1)) {
			consumeToken(self);
			consumeToken(self);
			return createIncDecStat(expr, -1);
		}
	}
	return false;
}

Vector *parseImplBlock(Parser *self, char *name, char *as) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
		consumeToken(self);

		Vector *v = createVector(VECTOR_EXPONENTIAL);

		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
				consumeToken(self);
				break;
			}

			FunctionDecl *func = parseFunctionDecl(self);
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

Impl *parseImpl(Parser *self) {
	if (checkTokenTypeAndContent(self, IDENTIFIER, IMPL_KEYWORD, 0)) {
		consumeToken(self);

		char *name = NULL;
		if (checkTokenType(self, IDENTIFIER, 0)) {
			name = consumeToken(self)->content;
		}
		else {
			parserError("Implementation expecting Type to implement, found: %s", peekAtTokenStream(self, 0)->content);
		}

		char *as = NULL;
		if (checkTokenTypeAndContent(self, IDENTIFIER, IMPL_AS_KEYWORD, 0)) {
			if (checkTokenType(self, IDENTIFIER, 1)) {
				consumeToken(self);
				as = consumeToken(self)->content;
			}
			else {
				parserError("Implementation expects identifier after `as` keyword, found: %s", peekAtTokenStream(self, 0)->content);
			}
		}

		Vector *implBlock = parseImplBlock(self, name, as);
		if (implBlock) {
			Impl *impl = createImpl(name, as);
			impl->funcs = implBlock;
			return impl;
		}
	}

	return false;
}

StructuredStatement *parseStructuredStatement(Parser *self) {
	MatchStat *match = parseMatchStat(self);
	if (match) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->matchStmt = match;
		stmt->type = MATCH_STAT_NODE;
		return stmt;
	}

	Block *block = parseBlock(self);
	if (block) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->block = block;
		stmt->type = BLOCK_NODE;
		return stmt;
	}

	IfStat *ifs = parseIfStat(self);
	if (ifs) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->ifStmt = ifs;
		stmt->type = IF_STAT_NODE;
		return stmt;
	}

	ForStat *fer = parseForStat(self);
	if (fer) {
		StructuredStatement *stmt = createStructuredStatement();
		stmt->forStmt = fer;
		stmt->type = FOR_STAT_NODE;
		return stmt;
	}

	return false;
}

UnstructuredStatement *parseUnstructuredStatement(Parser *self) {
	Impl *impl = parseImpl(self);
	if (impl) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->impl = impl;
		stmt->type = IMPL_NODE;
		return stmt;
	}

	LeaveStat *leave = parseLeaveStat(self);
	if (leave) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->leave = leave;
		stmt->type = LEAVE_STAT_NODE;
		return stmt;
	}

	Call *call = parseCall(self);
	if (call) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->call = call;
		stmt->type = FUNCTION_CALL_NODE;
		if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Function call expects a semi-colon, found: %s", peekAtTokenStream(self, 0)->content);
		}
		return stmt;
	}

	Declaration *decl = parseDeclaration(self);
	if (decl) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		stmt->decl = decl;
		stmt->type = DECLARATION_NODE;
		return stmt;
	}

	Expression *expr = parseExpression(self);
	if (expr) {
		UnstructuredStatement *stmt = createUnstructuredStatement();
		if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);
		}
		else {
			parserError("Expected semi-colon at the end of expression, found %s", peekAtTokenStream(self, 0)->content);
		}
		stmt->expr = expr;
		stmt->type = EXPR_STAT_NODE;
		return stmt;
	}

	return false;
}

Macro *parseMacro(Parser *self) {
	if (!checkTokenTypeAndContent(self, OPERATOR, "!", 0)) {
		return false;
	}	

	LinkerFlagMacro *linker = parseLinkerFlagMacro(self);
	if (linker) {
		Macro *stmt = createMacro();
		stmt->linker = linker;
		stmt->type = LINKER_FLAG_MACRO_NODE;
		return stmt;
	}
	
	UseMacro *use = parseUseMacro(self);
	if (use) {
		Macro *stmt = createMacro();
		stmt->use = use;
		stmt->type = USE_MACRO_NODE;
		return stmt;
	}

	return false;
}

Statement *parseStatement(Parser *self) {
	Macro *macro = parseMacro(self);
	if (macro) {
		Statement *stmt = createStatement();
		stmt->macro = macro;
		stmt->type = MACRO_NODE;
		return stmt;
	}

	StructuredStatement *strucStmt = parseStructuredStatement(self);
	if (strucStmt) {
		Statement *stmt = createStatement();
		stmt->structured = strucStmt;
		stmt->type = STRUCTURED_STATEMENT_NODE;
		return stmt;
	}

	UnstructuredStatement *unstrucStmt = parseUnstructuredStatement(self);
	if (unstrucStmt) {
		Statement *stmt = createStatement();
		stmt->unstructured = unstrucStmt;
		stmt->type = UNSTRUCTURED_STATEMENT_NODE;
		return stmt;
	}

	return false;
}

Block *parseBlock(Parser *self) {
	if (checkTokenTypeAndContent(self, OPERATOR, SINGLE_STATEMENT_OPERATOR, 0)) {
		consumeToken(self);

		Block *block = createBlock();
		if (block) {
			Statement *stat = parseStatement(self);
			if (stat) {
				pushBackItem(block->stmtList->stmts, stat);
			}
			block->singleStatementBlock = true;
			return block;
		}
	}
	else if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)) {
		consumeToken(self);

		Block *block = createBlock();
		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, "}", 0)) {
				consumeToken(self);
				break;
			}

			Statement *stat = parseStatement(self);
			if (stat) {
				pushBackItem(block->stmtList->stmts, stat);
			}
		}
		return block;
	}

	return false;
}

FunctionDecl *parseFunctionDecl(Parser *self) {
	FunctionSignature *signature = parseFunctionSignature(self);
	if (signature) {
		if (checkTokenTypeAndContent(self, SEPARATOR, "{", 0)
			|| checkTokenTypeAndContent(self, OPERATOR, SINGLE_STATEMENT_OPERATOR, 0)) {
			Block *block = parseBlock(self);
			if (block) {
				FunctionDecl *decl = createFunctionDecl();
				decl->signature = signature;
				decl->body = block;
				decl->prototype = false;
				return decl;
			}
		} 
		else if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
			consumeToken(self);

			FunctionDecl *decl = createFunctionDecl();
			decl->signature = signature;
			decl->prototype = true;
			return decl;
		}
		else {
			parserError("Function `%s` expects a block, or semi-colon, found: %s", signature->name, peekAtTokenStream(self, 0)->content);
		}
	}

	return false;
}

VariableDecl *parseVariableDecl(Parser *self) {
	bool mutable = false;
	if (checkTokenTypeAndContent(self, IDENTIFIER, MUT_KEYWORD, 0)) {
		consumeToken(self);
		mutable = true;
	}

	if (checkTokenType(self, IDENTIFIER, 0)
		&& checkTokenTypeAndContent(self, OPERATOR, ":", 1)) {
		char *var_name = consumeToken(self)->content;
		Expression *rhand = NULL;

		if (checkTokenTypeAndContent(self, OPERATOR, ":", 0)) {
			// check for type inference
			bool inferred = false;
			if (checkTokenTypeAndContent(self, OPERATOR, "=", 1)) {
				inferred = true;
			}

			// consume dat colon bby
			consumeToken(self);
			
			Type *type = NULL;
			if (!inferred) {
				type = parseType(self);
				if (!type) {
					parserError("Expected type in variable declaration `%s`, found: %s", var_name, peekAtTokenStream(self, 0)->content);
				}
			}

			// var decl
			if (checkTokenTypeAndContent(self, OPERATOR, "=", 0)) {
				consumeToken(self);

				rhand = parseExpression(self);
				if (rhand) {
					if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
						consumeToken(self);
					}
					else {
						parserError("Expected semi-colon at the end of variable declaration `%s`, found: %s", var_name, peekAtTokenStream(self, 0)->content);
					}
					VariableDecl *decl = createVariableDecl(type, var_name, mutable, rhand);
					decl->assigned = true;
					decl->inferred = inferred;
					return decl;
				}
			}
			// var definition
			else if (checkTokenTypeAndContent(self, SEPARATOR, ";", 0)) {
				consumeToken(self); // eat the semi colon!

				VariableDecl *decl = createVariableDecl(type, var_name, mutable, rhand);
				decl->assigned = false;
				decl->inferred = inferred;
				return decl;
			}
			else {
				parserError("Dangling variable declaration for `%s`, found: %s", var_name, peekAtTokenStream(self, 0)->content);
			}
		}
	}

	return false;
}

Declaration *parseDeclaration(Parser *self) {
	FunctionDecl *func = parseFunctionDecl(self);
	if (func) {
		Declaration *decl = createDeclaration();
		decl->funcDecl = func;
		decl->type = FUNCTION_DECL_NODE;
		return decl;
	}

	StructDecl *struc = parseStructDecl(self);
	if (struc) {
		Declaration *decl = createDeclaration();
		decl->structDecl = struc;
		decl->type = STRUCT_DECL_NODE;
		return decl;
	}

	EnumDecl *enumDecl = parseEnumDecl(self);
	if (enumDecl) {
		Declaration *decl = createDeclaration();
		decl->enumDecl = enumDecl;
		decl->type = ENUM_DECL_NODE;
		return decl;
	}

	VariableDecl *varDecl = parseVariableDecl(self);
	if (varDecl) {
		Declaration *decl = createDeclaration();
		decl->varDecl = varDecl;
		decl->type = VARIABLE_DECL_NODE;
		return decl;
	}

	return false;
}

int getTokenPrecedence(Parser *self) {
	Token *tok = peekAtTokenStream(self, 0);

	if (!isASCII(tok->content[0]))
		return -1;

	Precedence *prec = NULL;
	if (hashmap_get(self->binopPrecedence, tok->content, (void**) &prec) == MAP_MISSING) {
		verboseModeMessage("Precedence doesnt exist for %s\n", tok->content);
		return -1;
	}

	int tokenPrecedence = prec->prec;
	if (tokenPrecedence <= 0) {
		return -1;
	}
	return tokenPrecedence;
}

TypeName *parseTypeName(Parser *self) {
	if (checkTokenType(self, IDENTIFIER, 0)) {
		return createTypeName(consumeToken(self)->content);
	}
	return false;
}

Expression *parseExpression(Parser *self) {
	Expression *expr = parsePrimaryExpression(self);
	if (!expr)
		return false;

	if (isValidBinaryOp(peekAtTokenStream(self, 0)->content)) {
		return parseBinaryOperator(self, 0, expr);
	}
	return expr;
}


ArrayType *parseArrayType(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "[", 0) 
		&& checkTokenTypeAndContent(self, SEPARATOR, "]", 1)) {
		consumeToken(self); // eat the [
		consumeToken(self); // eat the ]

		// parse the type
		Type *type = parseType(self);
		if (type) {
			return createArrayType(type);
		}
	}

	return false;
}

PointerType *parsePointerType(Parser *self) {
	if (checkTokenTypeAndContent(self, OPERATOR, "^", 0)) {
		consumeToken(self);
		Type *type = parseType(self);
		if (type) {
			return createPointerType(type);
		}
	}
	return false;
}

TupleType *parseTupleType(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
		consumeToken(self);

		TupleType *tuple = createTupleType();

		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, ")", 0)) {
				consumeToken(self);
				break;
			}

			Type *type = parseType(self);
			if (type) {
				pushBackItem(tuple->types, type);
				if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
					consumeToken(self);
				}
			}
		}

		return tuple;
	}
	return false;
}

OptionType *parseOptionType(Parser *self) {
	if (checkTokenTypeAndContent(self, OPERATOR, "<", 0)) {
		consumeToken(self);

		Type *type = parseType(self);
		if (type) {
			if (checkTokenTypeAndContent(self, OPERATOR, ">", 0)) {
				consumeToken(self);
			}
			return createOptionType(type);
		}

		errorMessage("todo option type");
		return false;
	}

	return false;
}

TypeLit *parseTypeLit(Parser *self) {
	ArrayType *arr = parseArrayType(self);
	if (arr) {
		TypeLit *lit = createTypeLit();
		lit->arrayType = arr;
		lit->type = ARRAY_TYPE_NODE;
		return lit;
	}

	OptionType *option = parseOptionType(self);
	if (option) {
		TypeLit *lit = createTypeLit();
		lit->optionType = option;
		lit->type = OPTION_TYPE_NODE;
		return lit;
	}

	TupleType *tuple = parseTupleType(self);
	if (tuple) {
		TypeLit *lit = createTypeLit();
		lit->tupleType = tuple;
		lit->type = TUPLE_TYPE_NODE;
		return lit;
	}

	PointerType *pntr = parsePointerType(self);
	if (pntr) {
		TypeLit *lit = createTypeLit();
		lit->pointerType = pntr;
		lit->type = POINTER_TYPE_NODE;
		return lit;
	}

	return false;
}

UnaryExpr *parseUnaryExpr(Parser *self) {
	if (isUnaryOp(peekAtTokenStream(self, 0)->content)) {
		char *op = consumeToken(self)->content;
		Expression *prim = parsePrimaryExpression(self);
		if (prim) {
			UnaryExpr *res = createUnaryExpr();
			res->lhand = prim;
			res->unaryOp = op;
			return res;
		}
	}

	return false;
}

Expression *parseBinaryOperator(Parser *self, int precedence, Expression *lhand) {
	for (;;) {
		int tokenPrecedence = getTokenPrecedence(self);
		if (tokenPrecedence < precedence)
			return lhand;

		Token *tok = peekAtTokenStream(self, 0);
		if (!isValidBinaryOp(tok->content)) {
			parserError("Invalid binary operator in expression `%s`", tok->content);
			return false;
		}
		char *binaryOp = consumeToken(self)->content;

		Expression *rhand = parsePrimaryExpression(self);
		if (!rhand)
			return false;

		int nextPrec = getTokenPrecedence(self);
		if (tokenPrecedence < nextPrec) {
			rhand = parseBinaryOperator(self, tokenPrecedence + 1, rhand);
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

ArrayInitializer *parseArrayInitializer(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "[", 0)) {
		consumeToken(self);

		ArrayInitializer *arr = createArrayInitializer();

		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, "]", 0)) {
				consumeToken(self);
				break;
			}

			Expression *expr = parseExpression(self);
			if (expr) {
				pushBackItem(arr->values, expr);
				if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
					consumeToken(self);
				}
			}
		}

		return arr;
	}

	return false;
}

TupleExpr *parseTupleExpr(Parser *self) {
	if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
		consumeToken(self);

		TupleExpr *tuple = createTupleExpr();

		while (true) {
			if (checkTokenTypeAndContent(self, SEPARATOR, ")", 0)) {
				consumeToken(self);
				break;
			}

			Expression *expr = parseExpression(self);
			if (expr) {
				pushBackItem(tuple->values, expr);
				if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
					consumeToken(self);
				}
			}
		}	

		return tuple;
	}

	return false;
}

Expression *parsePrimaryExpression(Parser *self) {
	if (checkTokenType(self, IDENTIFIER, 0) && (checkTokenTypeAndContent(self, SEPARATOR, "(", 1))) {
		Call *call = parseCall(self);
		if (call) {
			Expression *expr = createExpression();
			expr->call = call;
			expr->exprType = FUNCTION_CALL_NODE;
			return expr;
		}
	}

	if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
		TupleExpr *tuple = parseTupleExpr(self);
		if (tuple) {
			Expression *expr = createExpression();
			expr->tupleExpr = tuple;
			expr->exprType = TUPLE_EXPR_NODE;
			return expr;
		}
	}

	// array initializer.
	if (checkTokenTypeAndContent(self, SEPARATOR, "[", 0)) {
		ArrayInitializer *arr = parseArrayInitializer(self);
		if (arr) {
			Expression *expr = createExpression();
			expr->arrayInitializer = arr;
			expr->exprType = ARRAY_INITIALIZER_NODE;
			return expr;
		}
	}

	if (checkTokenType(self, IDENTIFIER, 0) 
		&& checkTokenTypeAndContent(self, SEPARATOR, ".", 1)) {
		Vector *members = createVector(VECTOR_EXPONENTIAL);
		while (true) {
			if (checkTokenType(self, IDENTIFIER, 0)) {
				char *iden = consumeToken(self)->content;
				pushBackItem(members, iden);
				if (checkTokenTypeAndContent(self, SEPARATOR, ".", 0)) {
					consumeToken(self);
				}
				else {
					break;
				}
			}
		}
	}

	UnaryExpr *unary = parseUnaryExpr(self);
	if (unary) {
		Expression *expr = createExpression();
		expr->unary = unary;
		expr->exprType = UNARY_EXPR_NODE;
		return expr;
	}

	Type *type = parseType(self);
	if (type) {
		Expression *expr = createExpression();
		expr->type = type;
		expr->exprType = TYPE_NODE;
		return expr;
	}

	Literal *lit = parseLiteral(self);
	if (lit) {
		Expression *expr = createExpression();
		expr->lit = lit;
		expr->exprType = LITERAL_NODE;
		return expr;
	}

	return false;
}

Call *parseCall(Parser *self) {
	if (checkTokenType(self, IDENTIFIER, 0) && 
		(checkTokenTypeAndContent(self, SEPARATOR, "(", 1) || checkTokenTypeAndContent(self, SEPARATOR, ".", 1))) {

		Vector *idens = NULL;
		if (checkTokenType(self, IDENTIFIER, 0)) {
			idens = createVector(VECTOR_LINEAR);
			while(true) {
				if (checkTokenType(self, IDENTIFIER, 0)) {
					pushBackItem(idens, consumeToken(self)->content);
				}
				if (checkTokenTypeAndContent(self, SEPARATOR, ".", 0)) {
					consumeToken(self);
				}
				if (!checkTokenType(self, IDENTIFIER, 0) || !checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
					break;
				}
			}
		}

		if (checkTokenTypeAndContent(self, SEPARATOR, "(", 0)) {
			consumeToken(self);

			Call *call = createCall(idens);
			while (true) {
				if (checkTokenTypeAndContent(self, SEPARATOR, ")", 0)) {
					consumeToken(self);
					break;
				}

				Expression *expr = parseExpression(self);
				if (expr) {
					pushBackItem(call->arguments, expr);
					if (checkTokenTypeAndContent(self, SEPARATOR, ",", 0)) {
						if (checkTokenTypeAndContent(self, SEPARATOR, ")", 1)) {
							parserError("Trailing comma in function call `%s`", getVectorItem(call->callee, 0));
						}
						consumeToken(self);
					}
				}
			}
			return call;
		}

	}
	return false;
}

/** UTILITY */

int getLiteralType(Token *tok) {
	switch (tok->type) {
	case CHARACTER:
		return LITERAL_CHAR;
	case HEX:
		return LITERAL_HEX_NUMBER;
	case DECIMAL:
		return LITERAL_DECIMAL_NUMBER;
	case WHOLE_NUMBER:
		return LITERAL_WHOLE_NUMBER;
	case STRING:
		return LITERAL_STRING;
	default:
		errorMessage("Unknown literal `%s`", tok->content);
		return LITERAL_ERRORED;
	}
}

Token *consumeToken(Parser *self) {
	Token *tok = getVectorItem(self->tokenStream, self->tokenIndex++);
	verboseModeMessage("consumed token: %s, current token is %s", tok->content,
			peekAtTokenStream(self, 0)->content);
	if (tok->type == END_OF_FILE) {
		self->parsing = false;
	}
	return tok;
}

bool checkTokenType(Parser *self, int type, int ahead) {
	return peekAtTokenStream(self, ahead)->type == type;
}

bool checkTokenTypeAndContent(Parser *self, int type, char *content,
		int ahead) {
	return peekAtTokenStream(self, ahead)->type == type
			&& !strcmp(peekAtTokenStream(self, ahead)->content, content);
}

bool matchTokenType(Parser *self, int type, int ahead) {
	if (checkTokenType(self, type, ahead)) {
		consumeToken(self);
		return true;
	}
	return false;
}

bool matchTokenTypeAndContent(Parser *self, int type, char *content,
		int ahead) {
	if (checkTokenTypeAndContent(self, type, content, ahead)) {
		consumeToken(self);
		return true;
	}
	return false;
}

Token *peekAtTokenStream(Parser *self, int ahead) {
	if (self->tokenIndex + ahead > self->tokenStream->size) {
		verboseModeMessage("Attempting to peek at out of bounds token: %d/%d", ahead,
				self->tokenStream->size);
		self->parsing = false;
		return NULL;
	}
	return getVectorItem(self->tokenStream, self->tokenIndex + ahead);
}

bool isLiteral(Parser *self, int ahead) {
	Token *tok = peekAtTokenStream(self, ahead);
	return tok->type == STRING 
			|| tok->type == WHOLE_NUMBER 
			|| tok->type == CHARACTER
			|| tok->type == DECIMAL
			|| tok->type == HEX;
}

/** DRIVER */

void startParsingSourceFiles(Parser *self, Vector *sourceFiles) {
	for (int i = 0; i < sourceFiles->size; i++) {
		SourceFile *file = getVectorItem(sourceFiles, i);
		self->tokenStream = file->tokens;
		self->parseTree = createVector(VECTOR_EXPONENTIAL);
		self->tokenIndex = 0;
		self->parsing = true;

		parseTokenStream(self);

		file->ast = self->parseTree;
	}
}

bool isValidBinaryOp(char *tok) {
	int size = ARR_LEN(BINARY_OPS);
	for (int i = 0; i < size; i++) {
		if (!strcmp(tok, BINARY_OPS[i])) {
			return true;
		}
	}
	return false;
}

void parseTokenStream(Parser *self) {
	while (!checkTokenType(self, END_OF_FILE, 0)) {
		Statement *stmt = parseStatement(self);
		if (stmt) {
			pushBackItem(self->parseTree, stmt);
		}
	}
}
