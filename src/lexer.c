#include "lexer.h"

Lexer *createLexer(char *input) {
	Lexer *lexer = malloc(sizeof(*lexer));
	if (!lexer) {
		perror("malloc: lexer");
	}
	lexer->input = input;
	lexer->dot = 0;
	lexer->inputChar = input[lexer->dot];
	return lexer;
}

void nextChar(Lexer *lexer) {
	lexer->inputChar = lexer->input[++lexer->dot];
}

char *inputToZString(Lexer *lexer, int start, int length) {
	char *result;
	strncpy(result = malloc(length + 1), &lexer->input[start], length);
	result[length] = '\0';
	return result;
}

void skipLayoutAndComment(Lexer *lexer) {
	while (isLayout(lexer->inputChar)) {
		nextChar(lexer);
	}
	while (isCommentOpener(lexer->inputChar)) {
		nextChar(lexer);
		while (!isCommentCloser(lexer->inputChar)) {
			if (isEndOfInput(lexer->inputChar)) return;
			nextChar(lexer);
		}
		nextChar(lexer);
		while (isLayout(lexer->inputChar)) {
			nextChar(lexer);
		}
	}
}

void recognizeIdentifier(Lexer *lexer) {
	lexer->token.class = IDENTIFIER;
	nextChar(lexer);

	while (isLetterOrDigit(lexer->inputChar)) {
		nextChar(lexer);
	}
	while (isUnderscore(lexer->inputChar) && isLetterOrDigit(lexer->input[lexer->dot + 1])) {
		nextChar(lexer);
		while (isLetterOrDigit(lexer->inputChar)) {
			nextChar(lexer);
		}
	}
}

void recognizeInteger(Lexer *lexer) {
	lexer->token.class = INTEGER;
	nextChar(lexer);
	while (isDigit(lexer->inputChar)) {
		nextChar(lexer);
	}
}

void getNextToken(Lexer *lexer) {
	int startDot;

	skipLayoutAndComment(lexer);

	startDot = lexer->dot;
	if (isEndOfInput(lexer->inputChar)) {
		lexer->token.class = END_OF_FILE;
		lexer->token.repr = "<END_OF_FILE>";
		return;
	}
	if (isLetter(lexer->inputChar)) {
		recognizeIdentifier(lexer);
	}
	else if (isDigit(lexer->inputChar)) {
		recognizeInteger(lexer);
	}
	else if (isOperator(lexer->inputChar)) {
		lexer->token.class = OPERATOR;
		nextChar(lexer);
	}
	else if (isSeparator(lexer->inputChar)) {
		lexer->token.class = SEPARATOR;
		nextChar(lexer);
	}
	else {
		lexer->token.class = ERRORNEOUS;
		nextChar(lexer);
	}
	lexer->token.repr = inputToZString(lexer, startDot, lexer->dot - startDot);
}

void destroyLexer(Lexer *lexer) {
	free(lexer);
}