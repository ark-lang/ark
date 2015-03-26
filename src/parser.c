#include "parser.h"

Parser *createParser() {
	Parser *parser = safeMalloc(sizeof(*parser));
	parser->tokenStream = NULL;
	parser->tokenIndex = 0;
	parser->parsing = true;
	parser->failed = false;
	return parser;
}

Token *consumeToken(Parser *parser) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex++);
}

bool checkTokenType(Parser *parser, int type, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type;
}

bool checkTokenTypeAndContent(Parser *parser, int type, char *content, int ahead) {
	return peekAtTokenStream(parser, 0)->type == type && !strcmp(peekAtTokenStream(parser, 0)->content, content);
}

Token *peekAtTokenStream(Parser *parser, int ahead) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex + ahead);
}

void startParsingSourceFiles(Parser *parser, Vector *sourceFiles) {
	int i;
	for (i = 0; i < sourceFiles->size; i++) {
		SourceFile *file = getVectorItem(sourceFiles, i);
		parser->tokenStream = file->tokens;
		parser->parseTree = createVector();
		parser->tokenIndex = 0;
		parser->parsing = true;

		parseTokenStream(parser);

		file->ast = parser->parseTree;
	}
}

void parseTokenStream(Parser *parser) {
	while (parser->parsing) {
		Token *tok = getVectorItem(parser->tokenStream, parser->tokenIndex);

		switch (tok->type) {
			case IDENTIFIER: break;
			case END_OF_FILE: parser->parsing = false; break;
		}
	}
}

void destroyParser(Parser *parser) {
	free(parser);
	debugMessage("Destroyed parser");
}
