#include "token.h"

const char *TOKEN_TYPE_NAME[] = {
	"TOKEN_IDENTIFIER", "TOKEN_OPERATOR", "TOKEN_SEPARATOR", "TOKEN_NUMBER",
	"TOKEN_ERRORNEOUS", "TOKEN_STRING", "TOKEN_CHARACTER", "TOKEN_UNKNOWN",
	"TOKEN_END_OF_FILE",
};

Token *createToken(int lineNumber, int lineStart, int charNumber, int charEnd, sds fileName) {
	Token *tok = safeMalloc(sizeof(*tok));
	tok->type = TOKEN_UNKNOWN;
	tok->content = NULL;
	tok->lineNumber = lineNumber;
	tok->lineStart = lineStart;
	tok->charNumber = charNumber;
	tok->charEnd = charEnd;
	tok->fileName = fileName;
	return tok;
}

char* getTokenContext(Vector *stream, Token *tok) {
	sds result = sdsempty();

	for (int i = 0; i < stream->size; i++) {
		Token *tempTok = getVectorItem(stream, i);
		if (tempTok->lineNumber == tok->lineNumber) {
			result = sdscat(result, tempTok->content);
			result = sdscat(result, " ");
		}
	}
	return result;
}


void destroyToken(Token *token) {
	free(token);
}

const char *getTokenTypeName(TokenType type) {
	if (type >= NUM_TOKEN_TYPES)
		errorMessage("Tried to get the name of a non-existent token type: %d", type);

	return TOKEN_TYPE_NAME[type];
}
