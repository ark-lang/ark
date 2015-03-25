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

Token *consumeToken(Parser *parser);

Token *peekAtTokenStream(Parser *parser, int ahead);

void startParsingSourceFiles(Parser *parser, Vector *sourceFiles);

void parseTokenStream(Parser *parser);

void destroyParser(Parser *parser);

#endif // parser_H
