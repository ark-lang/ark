#include "parser.h"

Parser *createParser() {
	Parser *parser = safeMalloc(sizeof(*parser));
	parser->tokenStream = NULL;
	parser->tokenIndex = 0;
	parser->parsing = true;
	parser->exitOnError = false;
	return parser;
}

Token *consumeToken(Parser *parser) {
	return getVectorItem(parser->tokenStream, parser->tokenIndex++);
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

}

void destroyParser(Parser *parser) {
	free(parser);
	debugMessage("Destroyed parser");
}
