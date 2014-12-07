#include "lexer.h"

Lexer *lexerCreate(string input) {
	Lexer *lexer = malloc(sizeof(*lexer));
	if (!lexer) {
		perror("malloc: failed to allocate memory for lexer");
	}
	lexer->input = input;
	lexer->pos = 0;
	lexer->charIndex = input[lexer->pos];
	lexer->token.type = UNKNOWN;
	lexer->token.content = "";
	return lexer;
}

void lexerNextChar(Lexer *lexer) {
	lexer->charIndex = lexer->input[++lexer->pos];
}

string lexerFlushBuffer(Lexer *lexer, int start, int length) {
	string result;
	strncpy(result = malloc(length + 1), &lexer->input[start], length);
	if (!result) { perror("malloc: failed to allocate memory for buffer flush"); }
	result[length] = '\0';
	return result;
}

void lexerSkipLayoutAndComment(Lexer *lexer) {
	while (isLayout(lexer->charIndex)) {
		lexerNextChar(lexer);
	}
	while (isCommentOpener(lexer->charIndex)) {
		lexerNextChar(lexer);
		while (!isCommentCloser(lexer->charIndex)) {
			if (isEndOfInput(lexer->charIndex)) return;
			lexerNextChar(lexer);
		}
		lexerNextChar(lexer);
		while (isLayout(lexer->charIndex)) {
			lexerNextChar(lexer);
		}
	}
}

void lexerRecognizeIdentifier(Lexer *lexer) {
	lexer->token.type = IDENTIFIER;
	lexerNextChar(lexer);

	while (isLetterOrDigit(lexer->charIndex)) {
		lexerNextChar(lexer);
	}
	while (isUnderscore(lexer->charIndex) && isLetterOrDigit(lexerPeekAhead(lexer, 1))) {
		lexerNextChar(lexer);
		while (isLetterOrDigit(lexer->charIndex)) {
			lexerNextChar(lexer);
		}
	}
}

void lexerRecognizeInteger(Lexer *lexer) {
	lexer->token.type = INTEGER;
	lexerNextChar(lexer);
	while (isDigit(lexer->charIndex)) {
		lexerNextChar(lexer);
	}
}

char lexerPeekAhead(Lexer *lexer, int ahead) {
	return lexer->input[lexer->pos + ahead];
}

void lexerGetNextToken(Lexer *lexer) {
	int startPos;

	lexerSkipLayoutAndComment(lexer);

	startPos = lexer->pos;
	if (isEndOfInput(lexer->charIndex)) {
		lexer->token.type = END_OF_FILE;
		lexer->token.content = "<END_OF_FILE>";
		return;
	}
	if (isLetter(lexer->charIndex)) {
		lexerRecognizeIdentifier(lexer);
	}
	else if (isDigit(lexer->charIndex)) {
		lexerRecognizeInteger(lexer);
	}
	else if (isOperator(lexer->charIndex)) {
		lexer->token.type = OPERATOR;
		lexerNextChar(lexer);
	}
	else if (isSeparator(lexer->charIndex)) {
		lexer->token.type = SEPARATOR;
		lexerNextChar(lexer);
	}
	else {
		lexer->token.type = ERRORNEOUS;
		lexerNextChar(lexer);
	}
	lexer->token.content = lexerFlushBuffer(lexer, startPos, lexer->pos - startPos);
}

void lexerDestroy(Lexer *lexer) {
	free(lexer);
}