#include "lexer.h"

Lexer *createLexer() {
	Lexer *self = safeMalloc(sizeof(*self));
	self->inputLength = 0;
	self->pos = 0;
	self->currentChar = '\0';
	self->tokenStream = NULL;
	self->running = true;
	self->lineNumber = 1;
	self->charNumber = 1;
	self->failed = false;
	return self;
}

void startLexingFiles(Lexer *self, Vector *sourceFiles) {
	for (int i = 0; i < sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(sourceFiles, i);

		// reset everything
		self->inputLength = strlen(sourceFile->alloyFileContents);
		if (self->inputLength <= 0) {
			errorMessage("File `%s` is empty.", sourceFile->name);
			self->failed = true;
			return;
		}
		self->input = sourceFile->alloyFileContents;
		self->pos = 0;
		self->lineNumber = 1;
		self->charNumber = 1;
		self->currentChar = self->input[self->pos];
		self->tokenStream = createVector(VECTOR_EXPONENTIAL);
		self->running = true;

		// get all the tokens
		while (self->running) {
			getNextToken(self);
		}

		// set the tokens to the current
		// token stream of the file being
		// lexed.
		sourceFile->tokens = self->tokenStream;
	}
}

void consumeCharacter(Lexer *self) {
	if (self->pos > (int) self->inputLength) {
		errorMessage("reached end of input, pos(%d) ... len(%d)", self->pos, self->inputLength);
		return;
	}
	// stop consuming if we hit the end of the file
	if(isEndOfInput(self->currentChar)) {
        return;
	}
	else if(self->currentChar == '\n') {
		self->charNumber = 0;	// reset the char number back to zero
		self->lineNumber++;
	}
    
	self->currentChar = self->input[++self->pos];
	self->charNumber++;
}

sds extractToken(Lexer *self, int start, int length) {
	return sdsnewlen(&self->input[start], length);
}

void skipLayoutAndComments(Lexer *self) {
	while (isLayout(self->currentChar)) {
		consumeCharacter(self);
	}

	// consume a block comment and its contents
	if (self->currentChar == '/' && peekAhead(self, 1) == '*') {
		// consume new comment symbols
		consumeCharacter(self);
		consumeCharacter(self);

		while (true) {
			consumeCharacter(self);

			if (isEndOfInput(self->currentChar)) {
				errorMessage("unterminated block comment");
				return;
			}

			if (self->currentChar == '*' && peekAhead(self, 1) == '/') {
				// consume the comment symbols
				consumeCharacter(self);
				consumeCharacter(self);

				// eat layout stuff like space etc
				while (isLayout(self->currentChar)) {
					consumeCharacter(self);
				}
				break;
			}
		}
	}

	// consume a single line comment
	while ((self->currentChar == '/' && peekAhead(self, 1) == '/')) {
		consumeCharacter(self);	// eat the /
		consumeCharacter(self);	// eat the /

		while (!isCommentCloser(self->currentChar)) {
			if (isEndOfInput(self->currentChar)) return;
			consumeCharacter(self);
		}
		
		while (isLayout(self->currentChar)) {
			consumeCharacter(self);
		}
	}
}

void expectCharacter(Lexer *self, char c) {
	if (self->currentChar == c) {
		consumeCharacter(self);
	}
	else {
		printf("expected `%c` but found `%c`\n", c, self->currentChar);
		return;
	}
}

void recognizeEndOfInputToken(Lexer *self) {
	consumeCharacter(self);
	pushInitializedToken(self, END_OF_FILE, "<END_OF_FILE>");
}

void recognizeIdentifierToken(Lexer *self) {
	consumeCharacter(self);

	while (isLetterOrDigit(self->currentChar)) {
		consumeCharacter(self);
	}
	while (isUnderscore(self->currentChar) && isLetterOrDigit(peekAhead(self, 1))) {
		consumeCharacter(self);
		while (isLetterOrDigit(self->currentChar)) {
			consumeCharacter(self);
		}
	}

	pushToken(self, IDENTIFIER);
}

void recognizeHexToken(Lexer *self) {
	consumeCharacter(self);

	if (self->currentChar == 'x' || self->currentChar == 'X') {
		consumeCharacter(self);

		while (isHexChar(self->currentChar)) {
			consumeCharacter(self);
		}
	}

	pushToken(self, HEX);
}

void recognizeNumberToken(Lexer *self) {
	consumeCharacter(self);

	if (self->currentChar == '.') {
		consumeCharacter(self); // consume dot
		
		while (isDigit(self->currentChar)) {
			consumeCharacter(self);
		}

		pushToken(self, DECIMAL);
	}
	else {
		// it'll do 
		bool isDecimal = false;

		while (isDigit(self->currentChar)) {
			if (peekAhead(self, 1) == '.') {
				consumeCharacter(self);
				while (isDigit(self->currentChar)) {
					consumeCharacter(self);
				}
				isDecimal = true;
			}
			consumeCharacter(self);
		}
		
		pushToken(self, isDecimal ? DECIMAL : WHOLE_NUMBER);
	}
}

void recognizeStringToken(Lexer *self) {
	expectCharacter(self, '"');

	// just consume everthing
	while (!isString(self->currentChar)) {
		consumeCharacter(self);
	}

	expectCharacter(self, '"');

	pushToken(self, STRING);
}

void recognizeCharacterToken(Lexer *self) {
	expectCharacter(self, '\'');

	if (isLetterOrDigit(self->currentChar)) {
		consumeCharacter(self); // consume character		
	}
	else {
		printf("error: empty character constant\n");
		return;
	}

	expectCharacter(self, '\'');

	pushToken(self, CHARACTER);
}

void recognizeOperatorToken(Lexer *self) {
	// stop the annoying := treated as an operator
	// treat them as individual operators instead.
	if (self->currentChar == ':' && peekAhead(self, 1) == '=') {
		consumeCharacter(self);
	}
	else {
		consumeCharacter(self);

		// for double operators
		if (isOperator(self->currentChar)) {
			consumeCharacter(self);
		}
	}

	pushToken(self, OPERATOR);
}

void recognizeEndOfLineToken(Lexer *self) {
	consumeCharacter(self);
}

void recognizeSeparatorToken(Lexer *self) {
	consumeCharacter(self);
	pushToken(self, SEPARATOR);
}

void recognizeErroneousToken(Lexer *self) {
	consumeCharacter(self);
	pushToken(self, ERRORNEOUS);
}

/** pushes a token with no content */
void pushToken(Lexer *self, int type) {
	Token *tok = createToken(self->lineNumber, self->charNumber);
	tok->type = type;
	tok->content = extractToken(self, self->startPos, self->pos - self->startPos);
	pushBackItem(self->tokenStream, tok);
}

/** pushes a token with content */
void pushInitializedToken(Lexer *self, int type, char *content) {
	Token *tok = createToken(self->lineNumber, self->charNumber);
	tok->type = type;
	tok->content = content;
	pushBackItem(self->tokenStream, tok);
}

char peekAhead(Lexer *self, int offset) {
	return self->input[self->pos + offset];
}

void getNextToken(Lexer *self) {
	self->startPos = 0;
	skipLayoutAndComments(self);
	self->startPos = self->pos;

	if (isEndOfInput(self->currentChar)) {
		recognizeEndOfInputToken(self);
		self->running = false;	// stop lexing
		return;
	}
	if (isDigit(self->currentChar) || self->currentChar == '.') {
		// number
		recognizeNumberToken(self);
	}
	else if (self->currentChar == '0' && (peekAhead(self, 0) == 'x' || peekAhead(self, 0) == 'X')) {
		recognizeHexToken(self);
	}
	else if (isLetterOrDigit(self->currentChar) || self->currentChar == '_') {
		// ident
		recognizeIdentifierToken(self);
	}
	else if (isString(self->currentChar)) {
		// string
		recognizeStringToken(self);
	}
	else if (isCharacter(self->currentChar)) {
		// character
		recognizeCharacterToken(self);
	}
	else if (isOperator(self->currentChar)) {
		// operator
		recognizeOperatorToken(self);
	}
	else if (isEndOfLine(self->currentChar)) {
		recognizeEndOfLineToken(self);
	}
	else if (isSeparator(self->currentChar)) {
		// separator
		recognizeSeparatorToken(self);
	}
	else {
		// errorneous
		recognizeErroneousToken(self);
	}
}

void destroyLexer(Lexer *self) {
	if (self->tokenStream != NULL) {
		for (int i = 0; i < self->tokenStream->size; i++) {
			Token *tok = getVectorItem(self->tokenStream, i);
			// eof's content isnt malloc'd so free would give us some errors
			if (tok->type != END_OF_FILE) {
				verboseModeMessage("Freed `%s` token", tok->content);
				sdsfree(tok->content);
			}
			destroyToken(tok);
		}
	}

	verboseModeMessage("Destroyed Lexer.");
	free(self);
}
