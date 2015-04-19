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
	IDENTIFIER,			// A | a ... z | Z | 0 ... 9 | _
	OPERATOR,			// todo
	SEPARATOR,			// todo
	HEX,				// 0x | A | B | C | D | E | F | 0 ... 9
	DECIMAL,			// digit { digit } "." { digit }
	WHOLE_NUMBER,		// digit { digit }
	ERRORNEOUS,			// error'd token
	STRING,
	CHARACTER,
	UNKNOWN,			// we dont know it durr
	END_OF_FILE, 		// EOF
} TokenType;

// this is just for debugging
static const char* TOKEN_NAMES[] = {
	"IDENTIFIER",
	"OPERATOR",
	"SEPARATOR",
	"HEX",
	"NUMBER",
	"ERRORNEOUS",
	"STRING",
	"CHARACTER",
	"UNKNOWN",
	"END_OF_FILE",
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