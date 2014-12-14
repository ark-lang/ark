#include "lexer.h"

static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

TokenPosition *tokenPositionCreate() {
	TokenPosition *pos = malloc(sizeof(*pos));
	if (!pos) {
		perror("malloc: could not allocate memory for Token Position");
		exit(1);
	}
	return pos;
}

void tokenPositionDestroy(TokenPosition *pos) {
	if (pos != NULL) {
		free(pos);
		pos = NULL;
	}
}

Token *tokenCreate() {
	Token *token = malloc(sizeof(*token));
	if (!token) {
		perror("malloc: failed to allocate memory for token");
		exit(1);
	}
	return token;
}

const char* getTokenName(Token *tok) {
	return TOKEN_NAMES[tok->type];
}

void tokenDestroy(Token *token) {
	if (token != NULL) {
		if (token->pos != NULL) {
			tokenPositionDestroy(token->pos);
		}
		free(token);
		token = NULL;
	}
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
	char* result = malloc(sizeof(char) * (length + 1));
	if (!result) { 
		perror("malloc: failed to allocate memory for buffer flush"); 
		exit(1);
	}

	strncpy(result, &lexer->input[start], length);
	result[length] = '\0';
	
	return result;
}

void lexerSkipLayoutAndComment(Lexer *lexer) {
	while (isLayout(lexer->charIndex)) {
		lexerNextChar(lexer);
	}

	while (lexer->charIndex == '/' && lexerPeekAhead(lexer, 1) == '*') {
		lexerNextChar(lexer); 	// consume /
		lexerNextChar(lexer);	// consume *

		// eat stuff inside of comment block
		while (lexer->charIndex != '*' && lexerPeekAhead(lexer, 1) != '/') {
			lexerNextChar(lexer);
		}

		lexerNextChar(lexer);	// consume *
		lexerNextChar(lexer);	// consume /
		
		// eat up the spaces
		while (isLayout(lexer->charIndex)) {
			lexerNextChar(lexer);
		}
	}

	while (lexer->charIndex == '/' && lexerPeekAhead(lexer, 1) == '/') {
		lexerNextChar(lexer);	// eat the /
		lexerNextChar(lexer);	// eat the /

		while (!isCommentCloser(lexer->charIndex)) {
			if (isEndOfInput(lexer->charIndex)) return;
			lexerNextChar(lexer);
		}
		
		lexer->lineNumber++; // increment line number
		
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
		printf("%s:%d:%d: error: expected `%c` but found `%c`\n", "file", lexer->currentToken->pos->lineNumber, lexer->currentToken->pos->charNumber, c, lexer->charIndex);
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
		printf("%s:%d:%d: error: empty character constant\n", "file", lexer->currentToken->pos->lineNumber, lexer->currentToken->pos->charNumber);
		exit(1);
	}

	lexerExpectCharacter(lexer, '\'');
}

char lexerPeekAhead(Lexer *lexer, int ahead) {
	return lexer->input[lexer->pos + ahead];
}

void lexerUpdateTokenPos(Lexer *lexer) {
	// this is broken fix this shit
	lexer->currentToken->pos->lineNumber = lexer->lineNumber;
	lexer->currentToken->pos->charNumber = lexer->pos;
	lexer->currentToken->pos->fileName = "swag";
}

void lexerGetNextToken(Lexer *lexer) {
	int startPos;
	lexerSkipLayoutAndComment(lexer);
	startPos = lexer->pos;

	lexer->currentToken = tokenCreate();
	lexer->currentToken->pos = tokenPositionCreate();

	if (isEndOfInput(lexer->charIndex)) {
		lexer->currentToken->type = END_OF_FILE;
		lexer->currentToken->content = "<END_OF_FILE>";
		lexer->running = false;	// stop lexing
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
	else if (isEndOfLine(lexer->charIndex)) {
		lexer->lineNumber++;
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

	lexerUpdateTokenPos(lexer);
	lexer->currentToken->content = lexerFlushBuffer(lexer, startPos, lexer->pos - startPos);
	vectorPushBack(lexer->tokenStream, lexer->currentToken);
}

void lexerDestroy(Lexer *lexer) {
	if (lexer == NULL) return;
	free(lexer);
	lexer = NULL;
}
