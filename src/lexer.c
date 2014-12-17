#include "lexer.h"

static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

Token *createToken() {
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

void destroyToken(Token *token) {
	if (token != NULL) {
		free(token);
		token = NULL;
	}
}

Lexer *createLexer(char* input) {
	Lexer *lexer = malloc(sizeof(*lexer));
	if (!lexer) {
		perror("malloc: failed to allocate memory for lexer");
		exit(1);
	}
	lexer->input = input;
	lexer->pos = 0;
	lexer->charIndex = input[lexer->pos];
	lexer->tokenStream = createVector();
	lexer->running = true;
	lexer->lineNumber = 0;

	return lexer;
}

void consumeCharacter(Lexer *lexer) {
	lexer->charIndex = lexer->input[++lexer->pos];
}

char* flushBuffer(Lexer *lexer, int start, int length) {
	char* result = malloc(sizeof(char) * (length + 1));
	if (!result) { 
		perror("malloc: failed to allocate memory for buffer flush"); 
		exit(1);
	}

	strncpy(result, &lexer->input[start], length);
	result[length] = '\0';
	return result;
}

void skipLayoutAndComments(Lexer *lexer) {
	while (isLayout(lexer->charIndex)) {
		consumeCharacter(lexer);
	}

	// consume a block comment and its contents
	if (lexer->charIndex == '/' && peekAhead(lexer, 1) == '*') {
		// consume new comment symbols
		consumeCharacter(lexer);
		consumeCharacter(lexer);

		while (true) {
			consumeCharacter(lexer);
			if (lexer->charIndex == '*' && peekAhead(lexer, 1) == '/') {
				consumeCharacter(lexer);
				consumeCharacter(lexer);
				while (isLayout(lexer->charIndex)) {
					consumeCharacter(lexer);
				}
				break;
			}
		}
	}

	// consume a single line comment
	while ((lexer->charIndex == '/' && peekAhead(lexer, 1) == '/') || (lexer->charIndex == '#')) {
		consumeCharacter(lexer);	// eat the /
		consumeCharacter(lexer);	// eat the /

		while (!isCommentCloser(lexer->charIndex)) {
			if (isEndOfInput(lexer->charIndex)) return;
			consumeCharacter(lexer);
		}
		
		lexer->lineNumber++; // increment line number
		
		while (isLayout(lexer->charIndex)) {
			consumeCharacter(lexer);
		}
	}
}

void expectCharacter(Lexer *lexer, char c) {
	if (lexer->charIndex == c) {
		consumeCharacter(lexer);
	}
	else {
		printf("error: expected `%c` but found `%c`\n", c, lexer->charIndex);
		exit(1);
	}
}

void recognizeIdentifierToken(Lexer *lexer) {
	consumeCharacter(lexer);

	while (isLetterOrDigit(lexer->charIndex)) {
		consumeCharacter(lexer);
	}
	while (isUnderscore(lexer->charIndex) && isLetterOrDigit(peekAhead(lexer, 1))) {
		consumeCharacter(lexer);
		while (isLetterOrDigit(lexer->charIndex)) {
			consumeCharacter(lexer);
		}
	}
}

void recognizeNumberToken(Lexer *lexer) {
	consumeCharacter(lexer);
	if (lexer->charIndex == '.') {
		consumeCharacter(lexer); // consume dot
		while (isDigit(lexer->charIndex)) {
			consumeCharacter(lexer);
		}
	}

	while (isDigit(lexer->charIndex)) {
		if (peekAhead(lexer, 1) == '.') {
			consumeCharacter(lexer);
			while (isDigit(lexer->charIndex)) {
				consumeCharacter(lexer);
			}
		}
		consumeCharacter(lexer);
	}
}

void recognizeStringToken(Lexer *lexer) {
	expectCharacter(lexer, '"');

	// just consume everthing
	while (!isString(lexer->charIndex)) {
		consumeCharacter(lexer);
	}

	expectCharacter(lexer, '"');
}

void recognizeCharacterToken(Lexer *lexer) {
	expectCharacter(lexer, '\'');

	if (isLetterOrDigit(lexer->charIndex)) {
		consumeCharacter(lexer); // consume character		
	}
	else {
		printf("error: empty character constant\n");
		exit(1);
	}

	expectCharacter(lexer, '\'');
}

char peekAhead(Lexer *lexer, int ahead) {
	return lexer->input[lexer->pos + ahead];
}

void getNextToken(Lexer *lexer) {
	int startPos;
	skipLayoutAndComments(lexer);
	startPos = lexer->pos;

	lexer->currentToken = createToken();

	if (isEndOfInput(lexer->charIndex)) {
		lexer->currentToken->type = END_OF_FILE;
		lexer->currentToken->content = "<END_OF_FILE>";
		lexer->running = false;	// stop lexing
		pushBackVectorItem(lexer->tokenStream, lexer->currentToken);
		return;
	}
	if (isLetter(lexer->charIndex)) {
		lexer->currentToken->type = IDENTIFIER;
		recognizeIdentifierToken(lexer);
	}
	else if (isDigit(lexer->charIndex) || lexer->charIndex == '.') {
		lexer->currentToken->type = NUMBER;
		recognizeNumberToken(lexer);
	}
	else if (isString(lexer->charIndex)) {
		lexer->currentToken->type = STRING;
		recognizeStringToken(lexer);
	}
	else if (isCharacter(lexer->charIndex)) {
		lexer->currentToken->type = CHARACTER;
		recognizeCharacterToken(lexer);
	}
	else if (isOperator(lexer->charIndex)) {
		lexer->currentToken->type = OPERATOR;
		consumeCharacter(lexer);
	}
	else if (isEndOfLine(lexer->charIndex)) {
		lexer->lineNumber++;
		consumeCharacter(lexer);
	}
	else if (isSeparator(lexer->charIndex)) {
		lexer->currentToken->type = SEPARATOR;
		consumeCharacter(lexer);
	}
	else {
		lexer->currentToken->type = ERRORNEOUS;
		consumeCharacter(lexer);
	}

	lexer->currentToken->content = flushBuffer(lexer, startPos, lexer->pos - startPos);
	pushBackVectorItem(lexer->tokenStream, lexer->currentToken);
}

void destroyLexer(Lexer *lexer) {
	if (lexer == NULL) return;
	free(lexer);
	lexer = NULL;
}
