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

/**
 * Different data types
 */
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
 * Peek at the token that is {@ahead} tokens
 * away in the token stream
 * 
 * @param parser instance of parser
 * @param ahead how far ahead to peek
 * @return the Token peeking at
 */
Token *parserPeekAhead(Parser *parser, int ahead);

/**
 * Checks if the next token type is the same as the given
 * token type. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserExpectType(Parser *parser, TokenType type);

/**
 * Checks if the next tokens content is the same as the given
 * content. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserExpectContent(Parser *parser, char *content);

/**
 * Checks if the next token type is the same as the given
 * token type and the token content is the same as the given
 * content, if not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @param content content to match
 * @return the token we matched
 */
Token *parserExpectTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserMatchType(Parser *parser, TokenType type);

/**
 * Checks if the current tokens content is the same as the given
 * content. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserMatchContent(Parser *parser, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type and the token content is the same as the given
 * content, if not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @param content content to match
 * @return the token we matched
 */
Token *parserMatchTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * if the token at the given index is the same type as the given one
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token is the same type as the given one
 */
bool parserTokenType(Parser *parser, TokenType type, int ahead);

/**
 * if the token at the given index has the same content as the given
 * @param parser the parser instance
 * @param type the type to check
 * @param ahead how far away the token is
 * @return if the current token has the same content as the given
 */
bool parserTokenContent(Parser *parser, char* content, int ahead);

/**
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token has the same content as the given
 */
bool parserTokenTypeAndContent(Parser *parser, TokenType type, char* content, int ahead);

/**
 * Parses an expression: currently only parses a number!
 * 
 * @param parser the parser instance
 * @return the expression parsed
 */
Expression parserParseExpression(Parser *parser);

/**
 * Prints the type and content of the current token
 */
void printCurrentToken(Parser *parser);

/**
 * Parses an integer
 * 
 * @param param the integer to parse
 */
void parserParseInteger(Parser *parser);

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