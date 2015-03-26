#ifndef parser_H
#define parser_H

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <assert.h>

#include "lexer.h"
#include "util.h"
#include "vector.h"
#include "ast.h"

#define STRUCT_KEYWORD 		"struct"
#define MUT_KEYWORD 		"mut"

/**
 * parser contents
 */
typedef struct {
	Vector *tokenStream;	// the stream of tokens to parse
	Vector *parseTree;		// the AST created

	int tokenIndex;			// current token
	bool parsing;			// if we're parsing
	bool failed;			// if parsing failed
} Parser;

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

IdentifierList *parseIdentifierList(Parser *parser);

/** UTILITIES */

/**
 * Returns the literal type based on the token
 * passed
 *
 * @param tok the token to check
 */
LiteralType getLiteralType(Token *tok);

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
 * Peek at the token stream by @ahead amount
 *
 * @param parser the parser instance
 * @param ahead how many tokens to peek ahead
 */
Token *peekAtTokenStream(Parser *parser, int ahead);

/** DRIVERS */

/**
 * Start parsing the source files
 *
 * @param parser the parser instance
 * @param sourceFiles the files to parse
 */
void startParsingSourceFiles(Parser *parser, Vector *sourceFiles);

/**
 * Parse the token stream
 *
 * @param parser the parser instance
 */
void parseTokenStream(Parser *parser);

#endif // parser_H
