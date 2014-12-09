#ifndef PARSER_H
#define PARSER_H

#include <stdlib.h>

#include "lexer.h"
#include "util.h"
#include "vector.h"

/**
 * Parser contents
 */
typedef struct {
	Vector *tokenStream;
	Vector *parseTree;
	int tokenIndex;
	bool parsing;
} Parser;

typedef enum {
	INTEGER,
} DataType;

/**
 * Node for an uninitialized
 * Variable
 */
typedef struct {
	DataType type;		// type of data to store
	Token *name;		// name of the variable
} VariableDefineNode;

/**
 * Node for an Expression
 */
typedef struct s_Expression {
	char type;
	Token *value;
	
	struct s_Expression *leftHand;
	char operand;
	struct s_Expression *rightHand;
} Expression;

/** 
 * Node for a Variable being declared
 */
typedef struct {
	VariableDefineNode vdn;
	Expression *expr;
} VariableDeclareNode;

/**
 * Create a new Parser instance
 * 
 * @param tokenStream the token stream to parse
 * @return instance of Parser
 */
Parser *parserCreate(Vector *tokenStream);

/**
 * Advances to the next token
 * 
 * @param parser parser instance
 * @return the token we consumed
 */
Token *parserConsumeToken(Parser *parser);

/**
 * Check if next token is the given type
 */
Token *parserExpectType(Parser *parser, TokenType type);

/**
 * Peek at the token that is {@ahead} tokens
 * away in the token stream
 * 
 * @param parser instance of parser
 * @param ahead how far ahead to peek
 * @return the Token peeking at
 */
Token *parserPeekAhead(Parser *parser, int ahead);

/**
 * Start parsing
 *
 * @param parser parser to start parsing
 */
void parserStartParsing(Parser *parser);

/**
 * Destroy the given Parser
 * 
 * @param parser the parser to destroy
 */
void parserDestroy(Parser *parser);

#endif // PARSER_H