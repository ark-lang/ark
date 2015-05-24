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
	TOKEN_STRING,			  // "some_string"
	TOKEN_CHARACTER,		  // 'c'
	TOKEN_UNKNOWN,            // we dont know it durr
	TOKEN_END_OF_FILE,        // EOF
	NUM_TOKEN_TYPES,
} TokenType;

extern const char *TOKEN_TYPE_NAME[];

/**
 * Token properties:
 * type 		- the token type
 * content 		- the token content
 * line_number 	- the line number of the token
 * line_start	- the index in the input where this token's line starts
 * char_number 	- the column of the first char in this token
 * char_end		- the column of the last char in this token
 */
typedef struct {
	int type;
	sds content;
	int lineNumber;
	int lineStart;
	int charNumber;
	int charEnd;
	sds fileName;
} Token;

/**
 * Create a token
 *
 * @param lineNumber the line number of the token
 * @param lineStart  the index in the input where this token's line starts
 * @param charNumber the column of the first char in the token
 * @param charEnd	 the column of the last char in the token
 */
Token *createToken(int lineNumber, int lineStart, int charNumber, int charEnd, sds fileName);

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

/**
 * Get the name for a token type
 */
const char *getTokenTypeName(TokenType type);

#endif // __TOKEN_H
