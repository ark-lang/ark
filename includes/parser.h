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

/**
 * parser contents
 */
typedef struct {
	Vector *tokenStream;
	Vector *parseTree;

	int tokenIndex;
	bool parsing;
	bool exitOnError;
} Parser;

Parser *createParser();

void destroyParser(Parser *parser);

/** AST */

IdentifierList *parseIdentifierList(Parser *parser);

/** UTILITIES */

Token *consumeToken(Parser *parser);

bool checkTokenType(Parser *parser, int type, int ahead);

bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead);

Token *peekAtTokenStream(Parser *parser, int ahead);

/** DRIVERS */

void startParsingSourceFiles(Parser *parser, Vector *sourceFiles);

void parseTokenStream(Parser *parser);

#endif // parser_H
