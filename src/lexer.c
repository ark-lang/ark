#include "lexer.h"

Token *tokenCreate() {
	return malloc(sizeof(Token));
}

void tokenDestroy(Token *token) {
	free(token);
}

Lexer *lexerCreate(char* input) {
	Lexer *lexer = malloc(sizeof(*lexer));
	if (!lexer) {
		perror("malloc: failed to allocate memory for lexer");
		exit(1);
	}
	lexer->input = input;
	lexer->pos = 0;
	lexer->charIndex = input[lexer->pos];
	lexer->tokenStream = vectorCreate();
	lexer->running = true;
	return lexer;
}

void lexerNextChar(Lexer *lexer) {
	lexer->charIndex = lexer->input[++lexer->pos];
}

char* lexerFlushBuffer(Lexer *lexer, int start, int length) {
	char* result;
	strncpy(result = malloc(length + 1), &lexer->input[start], length);
	if (!result) { 
		perror("malloc: failed to allocate memory for buffer flush"); 
		exit(1);
	}
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

	Token *tok = tokenCreate();

	if (isEndOfInput(lexer->charIndex)) {
		tok->type = END_OF_FILE;
		tok->content = "<END_OF_FILE>";
		lexer->running = false;
		vectorPushBack(lexer->tokenStream, tok);
		return;
	}
	if (isLetter(lexer->charIndex)) {
		tok->type = IDENTIFIER;
		lexerRecognizeIdentifier(lexer);
	}
	else if (isDigit(lexer->charIndex)) {
		tok->type = INTEGER;
		lexerRecognizeInteger(lexer);
	}
	else if (isOperator(lexer->charIndex)) {
		tok->type = OPERATOR;
		lexerNextChar(lexer);
	}
	else if (isSeparator(lexer->charIndex)) {
		tok->type = SEPARATOR;
		lexerNextChar(lexer);
	}
	else {
		tok->type = ERRORNEOUS;
		lexerNextChar(lexer);
	}

	tok->content = lexerFlushBuffer(lexer, startPos, lexer->pos - startPos);
	vectorPushBack(lexer->tokenStream, tok);
}

void lexerDestroy(Lexer *lexer) {
	free(lexer);
}