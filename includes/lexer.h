#ifndef LEXER_H
#define LEXER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define END_OF_FILE	0
#define IDENTIFIER	1
#define INTEGER		2
#define OPERATOR 	3
#define SEPARATOR 	4
#define	ERRORNEOUS	5

#define isEndOfInput(ch)	((ch) == '\0')
#define	isLayout(ch)		(!isEndOfInput(ch) && (ch) <= ' ')
#define isCommentOpener(ch)	((ch) == '#')
#define isCommentCloser(ch)	((ch) == '\n')
#define isUpperLetter(ch)	('A' <= (ch) && (ch) <= 'Z')
#define isLowerLetter(ch)	('a' <= (ch) && (ch) <= 'z')
#define isLetter(ch)		(isUpperLetter(ch) || isLowerLetter(ch))
#define isDigit(ch)			('0' <= (ch) && (ch) <= '9')
#define isLetterOrDigit(ch)	(isLetter(ch) || isDigit(ch))
#define isUnderscore(ch)	((ch) == '_')
#define isOperator(ch)		(strchr("+-*/=><!~?:&%^", (ch)) != 0)
#define isSeparator(ch)		(strchr(" ;,.`@(){} ", (ch)) != 0)

typedef struct {
	char *fileName;
	int lineNumber;
	int charNumber;
} TokenPosition;

typedef struct {
	int class;
	char *repr;
	TokenPosition pos;
} TokenType;

typedef struct {
	TokenType token;
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

#endif // LEXER_H