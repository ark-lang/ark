#include "token.h"

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


void destroyToken(Token *token) {
	free(token);
}