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
	return lexer;
}

void startLexingFiles(Lexer *lexer, Vector *sourceFiles) {
	int i;
	for (i = 0; i < sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(sourceFiles, i);
		printf("lexing %s\n", sourceFile->fileName);

		lexer->inputLength = strlen(sourceFile->fileContents);
		lexer->input = sourceFile->fileContents;
		lexer->currentChar = lexer->input[lexer->pos];
		lexer->tokenStream = createVector();

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
	if (token) {
		free(token);
	}
}

void consumeCharacter(Lexer *lexer) {
	if (lexer->pos > lexer->inputLength) {
		errorMessage("reached end of input");
		destroyLexer(lexer);
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

char* extractToken(Lexer *lexer, int start, int length) {
	char* result = safeMalloc(sizeof(char) * (length + 1));
	strncpy(result, &lexer->input[start], length);
	result[length] = '\0';
	return result;
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
				destroyLexer(lexer);
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

void recognize_end_of_input_token(Lexer *lexer) {
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
		recognize_end_of_input_token(lexer);
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
	if (lexer) {
		int i;
		for (i = 0; i < lexer->tokenStream->size; i++) {
			Token *tok = getVectorItem(lexer->tokenStream, i);
			// eof's content isnt malloc'd so free would give us some errors
			if (tok->type != END_OF_FILE) {
				free(tok->content);
			}
			destroyToken(tok);
		}

		free(lexer);
	}
}

// the ugly functions can go down here

// this is the holy grail of messy, and needs a lot of work
// i'm really considering writing some kind of string library
// for the compiler...
char* getTokenContext(Vector *stream, Token *tok, bool colour_error_token) {
	int line_num = tok->lineNumber;
	int result_size = 128;
	int result_index = 0;
	char *result = malloc(sizeof(char) * (result_size + 1));
	if (!result) {
		perror("malloc: failed to malloc memory for token context");
		return NULL;
	}

	int i;
	for (i = 0; i < stream->size; i++) {
		Token *temp_tok = getVectorItem(stream, i);
		if (temp_tok->lineNumber == line_num) {
			size_t len = strlen(temp_tok->content);

			int j;
			for (j = 0; j < len; j++) {
				// just in case we need to realloc
				if (result_index >= result_size) {
					result_size *= 2;
					result = realloc(result, sizeof(char) * (result_size + 1));
					if (!result) {
						perror("failed to reallocate memory for token context");
						return NULL;
					}
				}
				result[result_index++] = temp_tok->content[j];
			}

			// add a space so everything is cleaner
			result[result_index++] = ' ';
		}
	}

	result[result_index++] = '\0';
	return result;
}

char* get_line_number_context(Vector *stream, int line_num) {
	int result_size = 128;
	int result_index = 0;
	char *result = malloc(sizeof(char) * (result_size + 1));
	if (!result) {
		perror("malloc: failed to malloc memory for line number context");
		return NULL;
	}

	int i;
	for (i = 0; i < stream->size; i++) {
		Token *tok = getVectorItem(stream, i);
		if (tok->lineNumber == line_num) {
			size_t len = strlen(tok->content);
			int j;
			for (j = 0; j < len; j++) {
				// just in case we need to realloc
				if (result_index >= result_size) {
					result_size *= 2;
					result = realloc(result, sizeof(char) * (result_size + 1));
					if (!result) {
						perror("failed to reallocate memory for line number context");
						return NULL;
					}
				}

				// add the char to the result
				result[result_index++] = tok->content[j];
			}

			// add a space so everything is cleaner
			result[result_index++] = ' ';
		}
	}

	result[result_index++] = '\0';
	return result;
}
