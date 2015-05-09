#ifndef __PARSER_H
#define __PARSER_H

/**
 * This parses all of the tokens fed to the Parser via
 * the Lexer into an Abstract Syntax Tree or AST. As of writing
 * this, it does not handle _many_ errors, so if something breaks
 * it could either be the parsers not implemented it, or the code
 * you're feeding it is incorrect. Of course, this will be changed,
 * but for now it assumes that all of the code you feed it is semantically
 * and syntactically correct.
 *
 * The parser is written so that the parse functions will return false
 * if the grammar is not correct, or if it fails, bear that in mind.
 */

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

/**
 * These are keywords that we use a lot in the parser code,
 * some of them need better names.
 */
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
#define ENUM_KEYWORD				"enum"
#define MATCH_KEYWORD				"match"
#define CONTINUE_KEYWORD			"continue"
#define RETURN_KEYWORD				"return"
#define BREAK_KEYWORD				"break"
#define LINKER_FLAG_KEYWORD			"linker_flag"
#define IMPL_KEYWORD				"impl"
#define IMPL_AS_KEYWORD				"as"
#define ELSE_KEYWORD 				"else"

/**
 * Various operators that are used in the parser, also used
 * for lookups in certain cases i.e operator precedence parsing.
 */
static const char *logOp[] = { "||", "&&" };
static const char *relOp[] = { "==", "!=", "<", "<=", ">", ">=" };
static const char *addOp[] = { "+", "-", "|", "^", "=" };
static const char *mulOp[] = { "+", "/", "%", "<<", ">>", "&" };
static const char *unaryOp[] = { "+", "-", "!", "^", "<", ">", "*", "&" };

/**
 * Various types of Literals, mostly for lookups
 */
typedef enum {
	LITERAL_WHOLE_NUMBER,
	LITERAL_DECIMAL_NUMBER,
	LITERAL_HEX_NUMBER,
	LITERAL_STRING,
	LITERAL_CHAR,
	LITERAL_ERRORED
} LiteralType;

/**
 * The data types that are supported by Alloy,
 * again mostly for lookups.
 */
typedef enum {
	INT_64_TYPE, INT_32_TYPE, INT_16_TYPE, INT_8_TYPE,
	UINT_64_TYPE, UINT_32_TYPE, UINT_16_TYPE, UINT_8_TYPE,
	FLOAT_64_TYPE, FLOAT_32_TYPE,
	INT_TYPE, BOOL_TYPE, CHAR_TYPE, VOID_TYPE, UNKNOWN_TYPE
} DataType;

/**
 * @return the datatype (enum) that corresponds to
 * the given string. Note this compares names, not values,
 * i.e. don't mistake this for type inferrence. It will just
 * match "i32" and return INT_32_TYPE for instance.
 */
int getTypeFromString(char *type);

/**
 * Our parser properties
 */
typedef struct {
	Vector *tokenStream;	// the stream of tokens to parse
	Vector *parseTree;		// the AST created

	map_t binopPrecedence;
	int tokenIndex;			// current token

	bool parsing;			// if we're parsing
	bool failed;			// if parsing failed
} Parser;

/**
 * Just a little hack to get an
 * integer on the heap so that our
 * hashmap will use it.
 */
typedef struct {
	int prec;
} Precedence;

/**
 * Create a new parser instance
 */
Parser *createParser();

/**
 * Destroy the parser and its resources
 * @param  parser the parser instance
 */
void destroyParser(Parser *parser);

/**
 * Destroys a parse tree
 * @param  vector containing Statement structs
 */
void destroyParseTree(Vector *self);

/** AST */

/**
 * Parses a literal, e.g String, Character, Number literals.
 * @param  parser the parser instance
 */
Literal *parseLiteral(Parser *parser);

/**
 * Parse a group of identifiers separated by a comma, e.g.
 * vedant, is, a, peasant
 * @param  parser the parser instance
 */
IdentifierList *parseIdentifierList(Parser *parser);

/**
 * Parses a type, namely PointerTypes and ArrayTypes.
 * @param  parser the parser instance
 */
Type *parseType(Parser *parser);

/**
 * Parses a Field Declaration, i.e a variable defined
 * in a structure.
 * @param  parser the parser instance
 */
FieldDecl *parseFieldDecl(Parser *parser);

/**
 * Parses a list of Field Declarations wrapped inside a
 * Block.
 * @param  parser the parser instance
 */
FieldDeclList *parseFieldDeclList(Parser *parser);

/**
 * Parses a structure declaration
 * @param  parser the parser instance
 */
StructDecl *parseStructDecl(Parser *parser);

/**
 * Parses a parameter section, i.e a parameter in a 
 * parameter list, for instance: a: int is a parameter
 * section
 * @param  parser the parser instance
 */
ParameterSection *parseParameterSection(Parser *parser);

/**
 * Parses a group of parameter sections that are wrapped in
 * parenthesis, for instance: (a: int, y: str, x: double, z: i32)
 * @param  parser the parser instance
 */
Parameters *parseParameters(Parser *parser);

/**
 * Parses a function signature
 * @param  parser the parser instance
 */
FunctionSignature *parseFunctionSignature(Parser *parser);

/**
 * Parses an else statement
 * TODO
 * @param  parser the parser instance
 */
ElseStat *parseElseStat(Parser *parser);

/**
 * Parses an if statement, including the else/ifs and elses.
 * @param  parser the parser instance
 */
IfStat *parseIfStat(Parser *parser);

/**
 * Parses the top level node for a for statement, which includes
 * Infinite For Loops, Indexed For Loops, and Conditional For Loops.
 *
 * @param  parser the parser instance
 */
ForStat *parseForStat(Parser *parser);

/**
 * Parses a match clause contained within a match statement, for instance:
 *
 *     0 -> statement,
 *
 * These clauses can also have braces, for instance:
 *
 *     0 {
 *         statement;
 *         statement;
 *     }
 *
 * Note that this also appends a break statement to the end if there is no
 * other leave statement specified in the following statement after the clause.
 *
 * @param  parser the parser instance
 */
MatchClause *parseMatchClause(Parser *parser);

/**
 * Parses a match statement
 * @param  parser the parser instance
 */
MatchStat *parseMatchStat(Parser *parser);

/**
 * Parses a continue statement
 * @param  parser the parser instance
 */
ContinueStat *parseContinueStat(Parser *parser);

/**
 * Parses a simple break statement, these do not take
 * any expressions, so they are somewhat redundant.
 * @param  parser the parser instance
 */
BreakStat *parseBreakStat(Parser *parser);

/**
 * Parses a return statement, and the _optional_ following
 * expression
 * @param  parser the parser instance
 */
ReturnStat *parseReturnStat(Parser *parser);

/**
 * Parses a top level leave statement, e.g Return, Break, Continue
 * @param  parser the parser instance
 */
LeaveStat *parseLeaveStat(Parser *parser);

/**
 * Parses an increment or decrement statement, note that this
 * is kinda stupid and should be removed.
 * TODO, FIXME ?
 * @param  parser the parser instance
 */
IncDecStat *parseIncDecStat(Parser *parser);

/**
 * Parses an implementation block, i.e a group of functions
 * @param  parser the parser instance
 * @param  name   the name of the implementation, or struct owner
 * @param  as     the name of the optional alias
 * @return        the impl block node.
 */
Vector *parseImplBlock(Parser *parser, char *name, char *as);

/**
 * Parses an implementation node
 * @param  parser the parser instance
 * @return        the impl node we parsed
 */
Impl *parseImpl(Parser *parser);

/**
 * Parses an assignment node, this should be handled
 * as an expression probably
 * @param  parser the parser instance
 */
Assignment *parseAssignment(Parser *parser);

/**
 * Parses a top level statement
 * @param  parser the parser instance
 */
Statement *parseStatement(Parser *parser);

/**
 * Parses a block of statements
 * @param  parser the parser instance
 */
Block *parseBlock(Parser *parser);

/**
 * Parses a function declaration, which includes the
 * signature and the block, or just a prototype
 * @param  parser the parser instance
 */
FunctionDecl *parseFunctionDecl(Parser *parser);

/**
 * Parses a variable declaration (and definition too, kind of
 * a misnomer).
 * @param  parser the parser instance
 */
VariableDecl *parseVariableDecl(Parser *parser);

/**
 * Parses a top level declaration, i.e function decl, var decl,
 * etc.
 * @param  parser the parser instance
 */
Declaration *parseDeclaration(Parser *parser);

/**
 * Parses a type name, or an identifier for an expression?
 * i.e x: int, x is the TypeName
 * @param  parser the parser instance
 */
TypeName *parseTypeName(Parser *parser);

/**
 * Parses an expression, if the next token is an operator
 */
Expression *parseExpression(Parser *parser);

/**
 * An array type, this is used in the signature, although
 * it's kind of outdated with the new syntax, but arrays
 * don't work yet anyways so fuck it.
 * FIXME
 * @param  parser the parser instance
 */
ArrayType *parseArrayType(Parser *parser);

/**
 * A pointer type, also kinda fucked. Mostly for dereferencing
 * a pointer, but that should be handled as a unary expression,
 * not like a type.
 * @param  parser the parser instance
 */
PointerType *parsePointerType(Parser *parser);

/**
 * A literal type, for instance a string, char, number, etc.
 * @param  parser [description]
 * @return        [description]
 */
TypeLit *parseTypeLit(Parser *parser);

/**
 * Parses a unary expression, e.g +5 -2, etc...
 * @param  parser the parser instance
 */
UnaryExpr *parseUnaryExpr(Parser *parser);

/**
 * Operator precedence happens here!
 * @param  parser     the parser instance
 * @param  precedence the precedence of the previous op, or zero
 * @param  lhand      the left hand (old expr)
 */
Expression *parseBinaryOperator(Parser *parser, int precedence, Expression *lhand);

/**
 * Parses a primary expression, this is the actuall expression parsing,
 * i.e it will parse unaries, etc. The parseExpression part handles the
 * operator precedence parsing calls...
 * @param  parser the parser instance
 */
Expression *parsePrimaryExpression(Parser *parser);

/**
 * This parses a function call, which is treated as an expression.
 */
Call *parseCall(Parser *parser);

/** UTILITIES */

/**
 * Creates a new Precedence to be stored in a hashmap,
 * this is a small workaround for allocating it on the 
 * heap cleanly.
 * @param  prec the precedence to create
 */
static inline Precedence *createPrecedence(int prec) {
	Precedence *result = safeMalloc(sizeof(*result));
	result->prec = prec;
	return result;
}

/*
 * Function passed to hashmap_iterate() to destroy data.
 */
int destroyHashmapItem(any_t __attribute__((unused)) passedData, any_t item);

/**
 * Destroys the given precedence
 */
static inline void destroyPrecedence(Precedence *prec) {
	free(prec);
}

/**
 * @return the token precedence of the current operator
 */
int getTokenPrecedence(Parser *parser);

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

/**
 * Matches the {ahead} token to the type, it will eat the token
 * if it matches the criteria.
 */
bool matchTokenType(Parser *parser, int type, int ahead);

/**
 * Matches the {ahead} token to the type and content, it will eat
 * the token if it matches the criteria.
 */
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
	for (int i = 0; i < ARR_LEN(relOp); i++) {
		if (!strcmp(val, relOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a arithmetic operator
 */
static inline bool isAddOp(char *val) {
	for (int i = 0; i < ARR_LEN(addOp); i++) {
		if (!strcmp(val, addOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a arithmetic operator
 */
static inline bool isMulOp(char *val) {
	for (int i = 0; i < ARR_LEN(mulOp); i++) {
		if (!strcmp(val, mulOp[i])) return true;
	}
	return false;
}

/**
 * @return if the given string is a unary operator
 */
static inline bool isUnaryOp(char *val) {
	for (int i = 0; i < ARR_LEN(unaryOp); i++) {
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

/**
 * @return if the given operator/operand is valid
 */
bool isValidBinaryOp(char *val);

/**
 * Parse the token stream
 *
 * @param parser the parser instance
 */
void parseTokenStream(Parser *parser);

#endif // __PARSER_H
