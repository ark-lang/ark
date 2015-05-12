#ifndef __LEXER_H
#define __LEXER_H

/**
 * This handles the Lexical Analysis stage of our compiler,
 * it's a fairly quick and messy implementation and should be
 * cleaned up. Also, we should consider supporting utf-8, however
 * this may be tricky...
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "token.h"
#include "util.h"
#include "vector.h"
#include "sourcefile.h"

/**
 * Properties of our Lexer.
 */
typedef struct {
	sds input;				// input to lex
	int pos;				// position in the input
	int currentChar;		// current character
	int lineNumber;			// current line number
	size_t inputLength;		// sizeof lexer input
	int charNumber;			// current character at line
	int startPos;			// keeps track of positions without comments
	bool running;			// if lexer is running 
	bool failed;			// if lexing failed
	sds buffer;				// temporary token
	Vector *tokenStream;	// where the tokens are stored
	sds fileName;
} Lexer;

/**
 * Retrieves the line that a token is on
 * @param  lexer 	the lexer instance
 * @param  line_num the number to get context of
 * @return       	the context as a string
 */
char* getLineNumberContext(Vector *stream, int lineNumber);

/**
 * Create an instance of the Lexer
 * 
 * @param input the input to lex
 * @return instance of Lexer
 */
Lexer *createLexer();

/**
 * Start lexing the files we feed to the alloy compiler
 *
 * @param lexer the lexer instance
 * @param sourceFiles the files to lex
 */
void startLexingFiles(Lexer *lexer, Vector *sourceFiles);

/**
 * Simple substring, basically extracts the token from
 * the lexers input from [start .. start + length]
 * 
 * @param lexer instance of lexer
 * @param start start of the input
 * @param length of the input
 * @return string cut from buffer
 */
char* extractToken(Lexer *lexer, int start, int length);

/**
 * Advance to the next character, consuming the
 * current one.
 * 
 * @param lexer instance of the lexer
 */
void consumeCharacter(Lexer *lexer);

/**
 * Skips layout characters, such as spaces,
 * and comments, which are denoted with the 
 * pound (#).
 * 
 * @param lexer the lexer instance
 */
void skipLayoutAndComments(Lexer *lexer);

/**
 * Checks if current character is the given character
 * otherwise throws an error
 * 
 * @param lexer the lexer instance
 */
void expectCharacter(Lexer *lexer, char c);

/**
 * Recognize an identifier
 * 
 * @param lexer the lexer instance
 */
void recognizeIdentifierToken(Lexer *lexer);

/**
 * Recognize an Integer
 * 
 * @param lexer the lexer instance
 */
void recognizeNumberToken(Lexer *lexer);

/**
 * Recognize a String
 * 
 * @param lexer the lexer instance
 */
void recognizeStringToken(Lexer *lexer);

/**
 * Recognize a Character
 * 
 * @param lexer the lexer instance
 */
void recognizeCharacterToken(Lexer *lexer);

/**
 * Recognizes the given operator and pushes it
 * @param lexer the lexer for access to the token stream
 */
void recognizeOperatorToken(Lexer *lexer);

/**
 * Recognizes the end of line token
 * @param lexer the lexer for access to the token stream
 */
void recognizeEndOfLineToken(Lexer *lexer);

/**
 * Recognizes a separator token and pushes it
 * to the tree
 * @param lexer the lexer for access to the token stream
 */
void recognizeSeparatorToken(Lexer *lexer);

/**
 * Recognizes an errored token and pushes it to the
 * tree
 * @param lexer the lexer for access to the token stream
 */
void recognizeErroneousToken(Lexer *lexer);

/**
 * Pushes a token to the token tree, also captures the 
 * token content so you don't have to.
 * 
 * @param lexer the lexer for access to the token tree
 * @param type  the type of token
 */
void pushToken(Lexer *lexer, int type);

/**
 * Pushes a token with content to the token tree
 * @param lexer   the lexer for access to the token tree
 * @param type    the type of token to push
 * @param content the content to push
 */
void pushInitializedToken(Lexer *lexer, int type, char *content);

/**
 * Peek ahead in the character stream by
 * the given amount
 * 
 * @lexer instance of lexer
 * @ahead amount to peek by
 * @return the char we peeked at
 */
char peekAhead(Lexer *lexer, int ahead);

/**
 * Process the next token in the token stream
 * 
 * @param lexer the lexer instance
 */
void getNextToken(Lexer *lexer);

/**
 * Destroys the given lexer instance,
 * freeing any memory
 * 
 * @param lexer the lexer instance to destroy
 */
void destroyLexer(Lexer *lexer);

/**
 * @return if the character given is the end of input
 * @param ch the character to check
 */
static inline bool isEndOfInput(char ch) { 
	return ch == '\0'; 
}

/**
 * @return if the character given is a layout character
 * @param ch the character to check
 */
static inline bool isLayout(char ch) { 
	return !isEndOfInput(ch) && (ch) <= ' '; 
}

/**
 * @return if the character given is a comment closer 
 * @param ch the character to check
 */
static inline bool isCommentCloser(char ch) { 
	return ch == '\n'; 
}

/**
 * @return if the character given is an uppercase letter
 * @param ch the character to check
 */
static inline bool isUpperLetter(char ch) { 
	return 'A' <= ch && ch <= 'Z'; 
}

/**
 * @return if the character given is a lower case letter
 * @param ch the character to check
 */
static inline bool isLowerLetter(char ch) { 
	return 'a' <= ch && ch <= 'z'; 
}

/**
 * @return if the character given is a letter a-z, A-Z
 * @param ch the character to check
 */
static inline bool isLetter(char ch) { 
	return isUpperLetter(ch) || isLowerLetter(ch); 
}

/**
 * @return if the character given is a digit 0-9
 * @param ch the character to check
 */
static inline bool isDigit(char ch) { 
	return '0' <= ch && ch <= '9'; 
}

/**
 * @return if the character given is a hex number,
 * i.e 0xABCDEF 0x0123456789
 */
static inline bool isHexChar(char ch) {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F');
}

/**
 * @return if the character given is a letter or digit a-z, A-Z, 0-9
 * @param ch the character to check
 */
static inline bool isLetterOrDigit(char ch) { 
	return isLetter(ch) || isDigit(ch); 
}

/**
 * @return if the character given is an underscore
 * @param ch the character to check
 */
static inline bool isUnderscore(char ch) { 
	return ch == '_'; 
}

/**
 * @return if the character given is a quote, denoting a string
 * @param ch the character to check
 */
static inline bool isString(char ch) { 
	return ch == '"'; 
}

/**
 * @return if the character given is a single quote, denoting a character
 * @param ch the character to check
 */
static inline bool isCharacter(char ch) { 
	return ch == '\''; 
}

/**
 * @return if the character given is an operator
 * @param ch the character to check
 */
static inline bool isOperator(char ch) { 
	return (strchr("+-*/=><!~?:|&%^\"'", ch) != 0); 
}

static inline bool isExpressionOperator(char ch) { 
	return (strchr("+-*/=><!~?:|&%^\"'()", ch) != 0); 
}

/**
 * @return if the character given is a separator
 * @param ch the character to check
 */
static inline bool isSeparator(char ch) { 
	return (strchr(" ;,.`@(){}[] ", ch) != 0); 
}

/**
 * @return if the character is end of line to track line number
 * @param ch character to check
 */
static inline bool isEndOfLine(char ch) { 
	return ch == '\n'; 
}

#endif // __LEXER_H
