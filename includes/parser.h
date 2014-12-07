#ifndef PARSER_H
#define PARSER_H

#include <stdlib.h>

#include "lexer.h"
#include "vector.h"

/**
 * Parser contents
 */
typedef struct {
	Vector *tokenStream;
} Parser;

/**
 * Create a new Parser instance
 * 
 * @param tokenStream the token stream to parse
 * @return instance of Parser
 */
Parser *parserCreate(Vector *tokenStream);

/**
 * Destroy the given Parser
 * 
 * @param parser the parser to destroy
 */
void parserDestroy(Parser *parser);

#endif // PARSER_H