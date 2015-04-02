#ifndef __TOKEN_H
#define __TOKEN_H

#include "util.h"
#include "vector.h"

/**
 * Different types of token
 */
typedef enum {
	END_OF_FILE, IDENTIFIER, NUMBER,
	OPERATOR, SEPARATOR, ERRORNEOUS,
	STRING, CHARACTER, UNKNOWN, SPECIAL_CHAR,
	SKIP_TOKEN
} TokenType;

// this is just for debugging
static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

/**
 * Token properties:
 * type 		- the token type
 * content 		- the token content
 * line_number 	- the line number of the token
 * char_number 	- the number of the char of the token
 */
typedef struct {
	int type;
	sds content;
	int lineNumber;
	int charNumber;
} Token;

/**
 * Create a token
 *
 * @param lineNumber the line number of the token
 * @param charNumber the character number
 */
Token *createToken(int lineNumber, int charNumber);

/**
 * Retrieve the token name of the given token
 */
const char* getTokenName(Token *tok);

/**
 * Retrieves the line that a token is on
 * @param  lexer              the lexer instance
 * @param  tok                the token to get the context of
 * @return                    the context as a string
 */
char* getTokenContext(Vector *stream, Token *tok);

/**
 * Destroy the token and its resources
 */
void destroyToken(Token *token);

#endif // __TOKEN_H