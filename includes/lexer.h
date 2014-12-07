#ifndef LEXER_H
#define LEXER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "vector.h"

/** Types of token */
typedef enum {
	END_OF_FILE, IDENTIFIER, INTEGER,
	OPERATOR, SEPARATOR, ERRORNEOUS,
	UNKNOWN
} TokenType;

/** Information on Token for debug */
typedef struct {
	char* fileName;
	int lineNumber;
	int charNumber;
} TokenPosition;

/** Properties of a Token or Lexeme */
typedef struct {
	int type;
	char* content;
	TokenPosition pos;
} Token;

Token *tokenCreate();

void tokenDestroy(Token *token);

/** Lexer stuff */
typedef struct {
	Vector *tokenStream;
	char* input;			// input to lex
	int pos;				// position in the input
	int charIndex;			// current character
	bool running;			// if lexer is running 
} Lexer;

/**
 * Create an instance of the Lexer
 * 
 * @param input the input to lex
 * @return instance of Lexer
 */
Lexer *lexerCreate(char* input);

/**
 * Simple substring implementation,
 * used to flush buffer. This is malloc'd memory, so free it!
 * 
 * @param lexer instance of lexer
 * @param start start of the input
 * @param length of the input
 * @return string cut from buffer
 */
char* lexerFlushBuffer(Lexer *lexer, int start, int length);

/**
 * Advance to the next character, consuming the
 * current one.
 * 
 * @param lexer instance of the lexer
 */
void lexerNextChar(Lexer *lexer);

/**
 * Skips layout characters, such as spaces,
 * and comments, which are denoted with the 
 * pound (#).
 * 
 * @param lexer the lexer instance
 */
void lexerSkipLayoutAndComment(Lexer *lexer);

/**
 * Recognize an identifier
 * 
 * @param lexer the lexer instance
 */
void lexerRecognizeIdentifier(Lexer *lexer);

/**
 * Recognize an Integer
 * 
 * @param lexer the lexer instance
 */
void lexerRecognizeInteger(Lexer *lexer);

/**
 * Peek ahead in the character stream by
 * the given amount
 * 
 * @lexer instance of lexer
 * @ahead amount to peek by
 * @return the char we peeked at
 */
char lexerPeekAhead(Lexer *lexer, int ahead);

/**
 * Process the next token in the token stream
 * 
 * @param lexer the lexer instance
 */
void lexerGetNextToken(Lexer *lexer);

/**
 * Destroys the given lexer instance,
 * freeing any memory
 * 
 * @param lexer the lexer instance to destroy
 */
void lexerDestroy(Lexer *lexer);

/**
 * @return if the character given is the end of input
 * @param ch the character to check
 */
static inline bool isEndOfInput(char ch) 		{ return(ch == '\0'); }

/**
 * @return if the character given is a comment opener (#)
 * @param ch the character to check
 */
static inline bool isCommentOpener(char ch) 	{ return(ch == '#'); }

/**
 * @return if the character given is a layout character
 * @param ch the character to check
 */
static inline bool isLayout(char ch) 			{ return(!isEndOfInput(ch) && (ch) <= ' '); }

/**
 * @return if the character given is a comment closer 
 * @param ch the character to check
 */
static inline bool isCommentCloser(char ch) 	{ return(ch == '\n'); }

/**
 * @return if the character given is an uppercase letter
 * @param ch the character to check
 */
static inline bool isUpperLetter(char ch) 		{ return('A' <= ch && ch <= 'Z'); }

/**
 * @return if the character given is a lower case letter
 * @param ch the character to check
 */
static inline bool isLowerLetter(char ch) 		{ return('a' <= ch && ch <= 'z'); }

/**
 * @return if the character given is a letter a-z, A-Z
 * @param ch the character to check
 */
static inline bool isLetter(char ch) 			{ return(isUpperLetter(ch) || isLowerLetter(ch)); }

/**
 * @return if the character given is a digit 0-9
 * @param ch the character to check
 */
static inline bool isDigit(char ch) 			{ return('0' <= ch && ch <= '9'); }

/**
 * @return if the character given is a letter or digit a-z, A-Z, 0-9
 * @param ch the character to check
 */
static inline bool isLetterOrDigit(char ch) 	{ return(isLetter(ch) || isDigit(ch)); }

/**
 * @return if the character given is an underscore
 * @param ch the character to check
 */
static inline bool isUnderscore(char ch) 		{ return(ch == '_'); }

/**
 * @return if the character given is an operator
 * @param ch the character to check
 */
static inline bool isOperator(char ch) 			{ return(strchr("+-*/=><!~?:&%^", ch) != 0); }

/**
 * @return if the character given is a separator
 * @param ch the character to check
 */
static inline bool isSeparator(char ch) 		{ return(strchr(" ;,.`@(){} ", ch) != 0); }

#endif // LEXER_H