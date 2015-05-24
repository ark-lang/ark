#ifndef __TOKEN_H
#define __TOKEN_H

/**
 * This represents a Token, a Token has a type,
 * a value that it contains (content), a lineNumber
 * and a charNumber. The last two are mostly for debugging
 * when we find errors in the program.
 */

#include "util.h"
#include "vector.h"

/**
 * Different types of token
 */
typedef enum {
	TOKEN_IDENTIFIER,         // A | a ... z | Z | 0 ... 9 | _
	TOKEN_OPERATOR,           // todo
	TOKEN_SEPARATOR,          // todo
	TOKEN_NUMBER,             // generic number (integer, floating, hex, bin, octal)
	TOKEN_ERRORNEOUS,         // error'd token
	TOKEN_STRING,
	TOKEN_CHARACTER,
	TOKEN_UNKNOWN,            // we dont know it durr
	TOKEN_END_OF_FILE,        // EOF
} TokenType;

/**
 * Token properties:
 * type 		- the token type
 * content 		- the token content
 * line_number 	- the line number of the token
 * charStart 	- the column of the first character in this token
 * charEnd		- the column of the last character in this token
 * inputStart	- the index in the input at which this token starts
 * inputEnd		- the index in the input at which this token ends
 */
typedef struct {
	int type;
	sds content;
	int lineNumber;
	int charStart;
	int charEnd;
	int inputStart;
	int inputEnd;
	sds fileName;
} Token;

/**
 * Create a token
 *
 * @param lineNumber the line number of the token
 * @param charNumber the character number
 */
Token *createToken(int lineNumber, int charStart, int charEnd, int inputStart, int inputEnd, sds fileName);

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
