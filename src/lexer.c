#include "lexer.h"

static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

Token *tokenCreate() {
	return malloc(sizeof(Token));
}

const char* getTokenName(Token *tok) {
	return TOKEN_NAMES[tok->type];
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
	lexer->lineNumber = 1;
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
		lexer->lineNumber++; // increment line number
		lexerNextChar(lexer);
		while (isLayout(lexer->charIndex)) {
			lexerNextChar(lexer);
		}
	}
}

void lexerExpectCharacter(Lexer *lexer, char c) {
	if (lexer->charIndex == c) {
		lexerNextChar(lexer);
	}
	else {
		printf("%s:%d:%d: error: expected `%c` but found `%c`\n", "file", lexer->lineNumber, lexer->currentToken->pos.charNumber, c, lexer->charIndex);
		exit(1);
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

void lexerRecognizeNumber(Lexer *lexer) {
	lexerNextChar(lexer);
	if (lexer->charIndex == '.') {
		lexerNextChar(lexer); // consume dot
		while (isDigit(lexer->charIndex)) {
			lexerNextChar(lexer);
		}
	}

	while (isDigit(lexer->charIndex)) {
		if (lexerPeekAhead(lexer, 1) == '.') {
			lexerNextChar(lexer);
			while (isDigit(lexer->charIndex)) {
				lexerNextChar(lexer);
			}
		}
		lexerNextChar(lexer);
	}
}

void lexerRecognizeString(Lexer *lexer) {
	lexerExpectCharacter(lexer, '"');

	// just consume everthing
	while (!isString(lexer->charIndex)) {
		lexerNextChar(lexer);
	}

	lexerExpectCharacter(lexer, '"');
}

void lexerRecognizeCharacter(Lexer *lexer) {
	lexerExpectCharacter(lexer, '\'');

	if (isLetterOrDigit(lexer->charIndex)) {
		lexerNextChar(lexer); // consume character		
	}
	else {
		printf("%s:%d:%d: error: empty character constant\n", "file", lexer->lineNumber, lexer->currentToken->pos.charNumber);
		exit(1);
	}

	lexerExpectCharacter(lexer, '\'');
}

char lexerPeekAhead(Lexer *lexer, int ahead) {
	char x = lexer->input[lexer->pos + ahead];
	return x;
}

void lexerGetNextToken(Lexer *lexer) {
	int startPos;
	lexerSkipLayoutAndComment(lexer);
	startPos = lexer->pos;

	lexer->currentToken = tokenCreate();
	lexer->currentToken->pos = (TokenPosition) {
		.fileName = "swag", .lineNumber = -1, .charNumber = lexer->pos
	};

	if (isEndOfInput(lexer->charIndex)) {
		lexer->currentToken->type = END_OF_FILE;
		lexer->currentToken->content = "<END_OF_FILE>";
		lexer->running = false;
		vectorPushBack(lexer->tokenStream, lexer->currentToken);
		return;
	}
	if (isLetter(lexer->charIndex)) {
		lexer->currentToken->type = IDENTIFIER;
		lexerRecognizeIdentifier(lexer);
	}
	else if (isDigit(lexer->charIndex) || lexer->charIndex == '.') {
		lexer->currentToken->type = NUMBER;
		lexerRecognizeNumber(lexer);
	}
	else if (isString(lexer->charIndex)) {
		lexer->currentToken->type = STRING;
		lexerRecognizeString(lexer);
	}
	else if (isCharacter(lexer->charIndex)) {
		lexer->currentToken->type = CHARACTER;
		lexerRecognizeCharacter(lexer);
	}
	else if (isOperator(lexer->charIndex)) {
		lexer->currentToken->type = OPERATOR;
		lexerNextChar(lexer);
	}
	else if (isSeparator(lexer->charIndex)) {
		lexer->currentToken->type = SEPARATOR;
		lexerNextChar(lexer);
	}
	else {
		lexer->currentToken->type = ERRORNEOUS;
		lexerNextChar(lexer);
	}

	lexer->currentToken->content = lexerFlushBuffer(lexer, startPos, lexer->pos - startPos);
	vectorPushBack(lexer->tokenStream, lexer->currentToken);
}

void lexerDestroy(Lexer *lexer) {
	free(lexer);
}
