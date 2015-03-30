#include "lexer.h"

Lexer *createLexer() {
	Lexer *lexer = safeMalloc(sizeof(*lexer));
	lexer->inputLength = 0;
	lexer->pos = 0;
	lexer->currentChar = '\0';
	lexer->tokenStream = NULL;
	lexer->running = true;
	lexer->lineNumber = 1;
	lexer->charNumber = 1;
	lexer->failed = false;
	return lexer;
}

void startLexingFiles(Lexer *lexer, Vector *sourceFiles) {
	int i;
	for (i = 0; i < sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(sourceFiles, i);

		// reset everything
		lexer->inputLength = strlen(sourceFile->alloyFileContents);
		if (lexer->inputLength <= 0) {
			errorMessage("File `%s` is empty.", sourceFile->name);
			lexer->failed = true;
			return;
		}
		lexer->input = sourceFile->alloyFileContents;
		lexer->pos = 0;
		lexer->lineNumber = 1;
		lexer->charNumber = 1;
		lexer->currentChar = lexer->input[lexer->pos];
		lexer->tokenStream = createVector(VECTOR_EXPONENTIAL);
		lexer->running = true;

		while (lexer->running) {
			getNextToken(lexer);
		}

		sourceFile->tokens = lexer->tokenStream;
	}
}

Token *createToken(int lineNumber, int charNumber) {
	Token *tok = safeMalloc(sizeof(*tok));
	tok->type = UNKNOWN;
	tok->content = NULL;
	tok->lineNumber = lineNumber;
	tok->charNumber = charNumber;
	return tok;
}

const char* getTokenName(Token *tok) {
	return TOKEN_NAMES[tok->type];
}

void destroyToken(Token *token) {
	free(token);
}

void consumeCharacter(Lexer *lexer) {
	if (lexer->pos > lexer->inputLength) {
		errorMessage("reached end of input, pos(%d) ... len(%d)", lexer->pos, lexer->inputLength);
		return;
	}
	// stop consuming if we hit the end of the file
	if(isEndOfInput(lexer->currentChar)) {
        return;
	}
	else if(lexer->currentChar == '\n') {
		lexer->charNumber = 0;	// reset the char number back to zero
		lexer->lineNumber++;
	}
    
	lexer->currentChar = lexer->input[++lexer->pos];
	lexer->charNumber++;
}

sds extractToken(Lexer *lexer, int start, int length) {
	return sdsnewlen(&lexer->input[start], length);
}

void skipLayoutAndComments(Lexer *lexer) {
	while (isLayout(lexer->currentChar)) {
		consumeCharacter(lexer);
	}

	// consume a block comment and its contents
	if (lexer->currentChar == '/' && peekAhead(lexer, 1) == '*') {
		// consume new comment symbols
		consumeCharacter(lexer);
		consumeCharacter(lexer);

		while (true) {
			consumeCharacter(lexer);

			if (isEndOfInput(lexer->currentChar)) {
				errorMessage("unterminated block comment");
				return;
			}

			if (lexer->currentChar == '*' && peekAhead(lexer, 1) == '/') {
				// consume the comment symbols
				consumeCharacter(lexer);
				consumeCharacter(lexer);

				// eat layout stuff like space etc
				while (isLayout(lexer->currentChar)) {
					consumeCharacter(lexer);
				}
				break;
			}
		}
	}

	// consume a single line comment
	while ((lexer->currentChar == '/' && peekAhead(lexer, 1) == '/')) {
		consumeCharacter(lexer);	// eat the /
		consumeCharacter(lexer);	// eat the /

		while (!isCommentCloser(lexer->currentChar)) {
			if (isEndOfInput(lexer->currentChar)) return;
			consumeCharacter(lexer);
		}
		
		while (isLayout(lexer->currentChar)) {
			consumeCharacter(lexer);
		}
	}
}

void expectCharacter(Lexer *lexer, char c) {
	if (lexer->currentChar == c) {
		consumeCharacter(lexer);
	}
	else {
		printf("expected `%c` but found `%c`\n", c, lexer->currentChar);
		return;
	}
}

void recognizeEndOfInputToken(Lexer *lexer) {
	consumeCharacter(lexer);
	pushInitializedToken(lexer, END_OF_FILE, "<END_OF_FILE>");
}

void recognizeIdentifierToken(Lexer *lexer) {
	consumeCharacter(lexer);

	while (isLetterOrDigit(lexer->currentChar)) {
		consumeCharacter(lexer);
	}
	while (isUnderscore(lexer->currentChar) && isLetterOrDigit(peekAhead(lexer, 1))) {
		consumeCharacter(lexer);
		while (isLetterOrDigit(lexer->currentChar)) {
			consumeCharacter(lexer);
		}
	}

	pushToken(lexer, IDENTIFIER);
}

void recognizeNumberToken(Lexer *lexer) {
	consumeCharacter(lexer);
	if (lexer->currentChar == '.') {
		consumeCharacter(lexer); // consume dot
		while (isDigit(lexer->currentChar)) {
			consumeCharacter(lexer);
		}
	}

	while (isDigit(lexer->currentChar)) {
		if (peekAhead(lexer, 1) == '.') {
			consumeCharacter(lexer);
			while (isDigit(lexer->currentChar)) {
				consumeCharacter(lexer);
			}
		}
		consumeCharacter(lexer);
	}

	pushToken(lexer, NUMBER);
}

void recognizeStringToken(Lexer *lexer) {
	expectCharacter(lexer, '"');

	// just consume everthing
	while (!isString(lexer->currentChar)) {
		consumeCharacter(lexer);
	}

	expectCharacter(lexer, '"');

	pushToken(lexer, STRING);
}

void recognizeCharacterToken(Lexer *lexer) {
	expectCharacter(lexer, '\'');

	if (isLetterOrDigit(lexer->currentChar)) {
		consumeCharacter(lexer); // consume character		
	}
	else {
		printf("error: empty character constant\n");
		return;
	}

	expectCharacter(lexer, '\'');

	pushToken(lexer, CHARACTER);
}

void recognizeOperatorToken(Lexer *lexer) {
	consumeCharacter(lexer);


	// for double operators
	if (isOperator(lexer->currentChar)) {
		consumeCharacter(lexer);
	}

	pushToken(lexer, OPERATOR);
}

void recognizeEndOfLineToken(Lexer *lexer) {
	consumeCharacter(lexer);
}

void recognizeSeparatorToken(Lexer *lexer) {
	consumeCharacter(lexer);
	pushToken(lexer, SEPARATOR);
}

void recognizeErroneousToken(Lexer *lexer) {
	consumeCharacter(lexer);
	pushToken(lexer, ERRORNEOUS);
}

/** pushes a token with no content */
void pushToken(Lexer *lexer, int type) {
	Token *tok = createToken(lexer->lineNumber, lexer->charNumber);
	tok->type = type;
	tok->content = extractToken(lexer, lexer->startPos, lexer->pos - lexer->startPos);
	pushBackItem(lexer->tokenStream, tok);
}

/** pushes a token with content */
void pushInitializedToken(Lexer *lexer, int type, char *content) {
	Token *tok = createToken(lexer->lineNumber, lexer->charNumber);
	tok->type = type;
	tok->content = content;
	pushBackItem(lexer->tokenStream, tok);
}

char peekAhead(Lexer *lexer, int offset) {
	return lexer->input[lexer->pos + offset];
}

void getNextToken(Lexer *lexer) {
	lexer->startPos = 0;
	skipLayoutAndComments(lexer);
	lexer->startPos = lexer->pos;

	if (isEndOfInput(lexer->currentChar)) {
		recognizeEndOfInputToken(lexer);
		lexer->running = false;	// stop lexing
		return;
	}
	if (isDigit(lexer->currentChar) || lexer->currentChar == '.') {
		// number
		recognizeNumberToken(lexer);
	}
	else if (isLetterOrDigit(lexer->currentChar) || lexer->currentChar == '_') {
		// ident
		recognizeIdentifierToken(lexer);
	}
	else if (isString(lexer->currentChar)) {
		// string
		recognizeStringToken(lexer);
	}
	else if (isCharacter(lexer->currentChar)) {
		// character
		recognizeCharacterToken(lexer);
	}
	else if (isOperator(lexer->currentChar)) {
		// operator
		recognizeOperatorToken(lexer);
	}
	else if (isEndOfLine(lexer->currentChar)) {
		recognizeEndOfLineToken(lexer);
	}
	else if (isSeparator(lexer->currentChar)) {
		// separator
		recognizeSeparatorToken(lexer);
	}
	else {
		// errorneous
		recognizeErroneousToken(lexer);
	}
}

void destroyLexer(Lexer *lexer) {
	if (lexer->tokenStream != NULL) {
		int i;
		for (i = 0; i < lexer->tokenStream->size; i++) {
			Token *tok = getVectorItem(lexer->tokenStream, i);
			// eof's content isnt malloc'd so free would give us some errors
			if (tok->type != END_OF_FILE) {
				debugMessage("Freed `%s` token", tok->content);
				sdsfree(tok->content);
			}
			destroyToken(tok);
		}
	}

	debugMessage("Destroyed Lexer.");
	free(lexer);
}

char* getTokenContext(Vector *stream, Token *tok) {
	sds result = sdsempty();
	
	int i;
	for (i = 0; i < stream->size; i++) {
		Token *tempTok = getVectorItem(stream, i);
		if (tempTok->lineNumber == tok->lineNumber) {
			result = sdscat(result, tempTok->content);
			result = sdscat(result, " ");
		}
	}
	return result;
}
