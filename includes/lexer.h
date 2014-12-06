#ifndef LEXER_H
#define LEXER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"

#define TOKEN_LIST_MAX_SIZE 512

typedef enum {
	END_OF_FILE,
	IDENTIFIER,
	INTEGER,
	OPERATOR,
	SEPARATOR,
	ERRORNEOUS,
	UNKNOWN
} TokenType;

typedef struct {
	char *fileName;
	int lineNumber;
	int charNumber;
} TokenPosition;

typedef struct {
	int class;
	char *repr;
	TokenPosition pos;
} Token;

typedef struct {
	Token token;
	char *input;
	int dot;
	int inputChar;
} Lexer;

Lexer *createLexer(char *input);

char *inputToZString(Lexer *lexer, int start, int length);

void nextChar(Lexer *lexer);

void skipLayoutAndComment(Lexer *lexer);

void recognizeIdentifier(Lexer *lexer);

void recognizeInteger(Lexer *lexer);

void getNextToken(Lexer *lexer);

void destroyLexer(Lexer *lexer);

static inline bool isEndOfInput(char ch)	{ 
	return(ch == '\0'); 
}

static inline bool isCommentOpener(char ch) { 
	return(ch == '#'); 
}

static inline bool isLayout(char ch)	{ 
	return(!isEndOfInput(ch) && (ch) <= ' '); 
}

static inline bool isCommentCloser(char ch) { 
	return(ch == '\n'); 
}

static inline bool isUpperLetter(char ch) { 
	return('A' <= ch && ch <= 'Z'); }

static inline bool isLowerLetter(char ch) { 
	return('a' <= ch && ch <= 'z'); 
}

static inline bool isLetter(char ch) { 
	return(isUpperLetter(ch) || isLowerLetter(ch)); 
}

static inline bool isDigit(char ch) { 
	return('0' <= ch && ch <= '9'); 
}

static inline bool isLetterOrDigit(char ch) { 
	return(isLetter(ch) || isDigit(ch)); 
}

static inline bool isUnderscore(char ch) { 
	return(ch == '_'); }

static inline bool isOperator(char ch) { 
return(strchr("+-*/=><!~?:&%^", ch) != 0); 
}

static inline bool isSeparator(char ch) { 
	return(strchr(" ;,.`@(){} ", ch) != 0); 
}


#endif // LEXER_H