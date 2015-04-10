#ifndef __PARSER_H
#define __PARSER_H

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <assert.h>

#include "lexer.h"
#include "util.h"
#include "vector.h"
#include "hashmap.h"
#include "ast.h"
#include "stack.h"
#include "scope.h"

#define SET_KEYWORD					"set"
#define VOID_KEYWORD				"void"
#define STRUCT_KEYWORD 				"struct"
#define USE_KEYWORD					"use"
#define MUT_KEYWORD 				"mut"
#define FUNCTION_KEYWORD 			"fn"
#define SINGLE_STATEMENT_OPERATOR 	"->"
#define IF_KEYWORD					"if"
#define ELSE_KEYWORD				"else"
#define FOR_KEYWORD					"for"
#define MATCH_KEYWORD				"match"
#define CONTINUE_KEYWORD			"continue"
#define RETURN_KEYWORD				"return"
#define BREAK_KEYWORD				"break"

static const char *logOp[] = { "||", "&&" };
static const char *relOp[] = { "==", "!=", "<", "<=", ">", ">=" };
static const char *addOp[] = { "+", "-", "|", "^" };
static const char *mulOp[] = { "+", "/", "%", "<<", ">>", "&" };
static const char *unaryOp[] = { "+", "-", "!", "^", "<", ">", "*", "&" };

typedef enum {
	LITERAL_NUMBER,
	LITERAL_STRING,
	LITERAL_CHAR,
	LITERAL_ERRORED
} LiteralType;

typedef enum {
	INT_64_TYPE, INT_32_TYPE, INT_16_TYPE, INT_8_TYPE,
	UINT_64_TYPE, UINT_32_TYPE, UINT_16_TYPE, UINT_8_TYPE,
	FLOAT_64_TYPE, FLOAT_32_TYPE,
	INT_TYPE, BOOL_TYPE, CHAR_TYPE, VOID_TYPE, UNKNOWN_TYPE
} DataType;

int getTypeFromString(char *type);

/**
 * parser contents
 */
typedef struct {
	Vector *tokenStream;	// the stream of tokens to parse
	Vector *parseTree;		// the AST created
	Stack *scope;

	map_t binopPrecedence;
	int tokenIndex;			// current token

	bool parsing;			// if we're parsing
	bool failed;			// if parsing failed
} Parser;

typedef struct {
	int prec;
} Precedence;

/**
 * Create the parser
 */
Parser *createParser();

/**
 * Destroy the parser and its resources
 *
 * @param parser the parser instance to destroy
 */
void destroyParser(Parser *parser);

/** AST */

Literal *parseLiteral(Parser *parser);

IdentifierList *parseIdentifierList(Parser *parser);

Type *parseType(Parser *parser);

FieldDecl *parseFieldDecl(Parser *parser);

FieldDeclList *parseFieldDeclList(Parser *parser);

StructDecl *parseStructDecl(Parser *parser);

ParameterSection *parseParameterSection(Parser *parser);

Parameters *parseParameters(Parser *parser);

Receiver *parseReceiver(Parser *parser);

FunctionSignature *parseFunctionSignature(Parser *parser);

ElseStat *parseElseStat(Parser *parser);

IfStat *parseIfStat(Parser *parser);

ForStat *parseForStat(Parser *parser);

MatchClause *parseMatchClause(Parser *parser);

MatchStat *parseMatchStat(Parser *parser);

Statement *parseStatement(Parser *parser);

Block *parseBlock(Parser *parser);

FunctionDecl *parseFunctionDecl(Parser *parser);

VariableDecl *parseVariableDecl(Parser *parser);

Declaration *parseDeclaration(Parser *parser);

BaseType *parseBaseType(Parser *parser);

TypeName *parseTypeName(Parser *parser);

Expression *parseExpression(Parser *parser);

ArrayType *parseArrayType(Parser *parser);

PointerType *parsePointerType(Parser *parser);

TypeLit *parseTypeLit(Parser *parser);

UnaryExpr *parseUnaryExpr(Parser *parser);

Expression *parseBinaryOperator(Parser *parser, int precedence, Expression *lhand);

Expression *parsePrimaryExpression(Parser *parser);

Call *parseCall(Parser *parser);

/** UTILITIES */

static inline Precedence *createPrecedence(int prec) {
	Precedence *result = safeMalloc(sizeof(*result));
	result->prec = prec;
	return result;
}

static inline void destroyPrecedence(Precedence *prec) {
	free(prec);
}

int getTokenPrecedence(Parser *parser);

void pushScope(Parser *parser);

void pushPointer(Parser *parser, char *name);

void popScope(Parser *parser, Block *block);

/**
 * Returns the literal type based on the token
 * passed
 *
 * @param tok the token to check
 */
int getLiteralType(Token *tok);

/**
 * Consumes the current token
 *
 * @param parser the parser instance
 */
Token *consumeToken(Parser *parser);

/**
 * Check if the tokens type is the same as the given
 *
 * @param parser the parser instance
 * @param type the type of token
 * @param ahead how many tokens to peek ahead
 */
bool checkTokenType(Parser *parser, int type, int ahead);

/**
 * Check if the tokens type and content is the same as given
 *
 * @param parser the parser instance
 * @param type the type of token to check
 * @param content the content to compare with
 * @param ahead how many tokens to peek ahead
 */
bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead);

bool matchTokenType(Parser *parser, int type, int ahead);

bool matchTokenTypeAndContent(Parser *parser, int type, char *content, int ahead);

/**
 * Peek at the token stream by @ahead amount
 *
 * @param parser the parser instance
 * @param ahead how many tokens to peek ahead
 */
Token *peekAtTokenStream(Parser *parser, int ahead);

/**
 * Checks if the token ahed is a literal
 */
bool isLiteral(Parser *parser, int ahead);

// whatever

/**
 * @return if the given string is a logical operator
 */
static inline bool isLogOp(char *val) {
	for (int i = 0; i < ARR_LEN(logOp); i++) {
		if (!strcmp(val, logOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a relational operator
 */
static inline bool isRelOp(char *val) {
	for (int i = 0; i < ARR_LEN(logOp); i++) {
		if (!strcmp(val, relOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a arithmetic operator
 */
static inline bool isAddOp(char *val) {
	for (int i = 0; i < ARR_LEN(logOp); i++) {
		if (!strcmp(val, addOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a arithmetic operator
 */
static inline bool isMulOp(char *val) {
	for (int i = 0; i < ARR_LEN(logOp); i++) {
		if (!strcmp(val, mulOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a unary operator
 */
static inline bool isUnaryOp(char *val) {
	for (int i = 0; i < ARR_LEN(logOp); i++) {
		if (!strcmp(val, unaryOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a binary operator
 */
static inline bool isBinaryOp(char *val) {
	return isLogOp(val) || isRelOp(val) || isAddOp(val) || isMulOp(val);
}

/** DRIVERS */

/**
 * Start parsing the source files
 *
 * @param parser the parser instance
 * @param sourceFiles the files to parse
 */
void startParsingSourceFiles(Parser *parser, Vector *sourceFiles);

bool isValidBinaryOp(char *val);

/**
 * Parse the token stream
 *
 * @param parser the parser instance
 */
void parseTokenStream(Parser *parser);

#endif // __PARSER_H
