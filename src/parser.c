#include "parser.h"

Parser *createParser(TokenType *tokens, int tokenListSize) {
	Parser *parser = malloc(sizeof(*parser));
	parser->tokens = tokens;
	parser->tokenListSize = tokenListSize;
	parser->currentTokenIndex = 0;
	return parser;
}

Expression *createExpression(Parser *parser) {
	Expression *expr = malloc(sizeof(*expr));
	return expr;
}

void destroyExpression(Parser *parser, Expression *expr) {
	free(expr);
}

bool parseOperator(Parser *parser, Operator *oper) {
	if (!strcmp(parser->tokens[parser->currentTokenIndex].repr, "+")) {
		*oper = '+';
		parser->currentTokenIndex += 1;
		return true;
	}
	if (!strcmp(parser->tokens[parser->currentTokenIndex].repr, "*")) {
		*oper = '*';
		parser->currentTokenIndex += 1;
		return true;
	}
	return false;
}

bool parseExpression(Parser *parser, Expression **expr) {
	Expression *newExpr = *expr = createExpression(parser);
	if (parser->tokens[parser->currentTokenIndex].class == INTEGER) {
		newExpr->type = 'D';
		newExpr->value = ((char) parser->tokens[parser->currentTokenIndex].repr[0]) - '0';
		parser->currentTokenIndex += 1;
		return true;
	}

	if (parser->tokens[parser->currentTokenIndex].class == SEPARATOR
		&& parser->tokens[parser->currentTokenIndex].repr[0] == '(') {
		newExpr->type = 'P';
		parser->currentTokenIndex += 1;

		if (!parseExpression(parser, &newExpr->left)) {
			const char *found = parser->tokens[parser->currentTokenIndex].repr;
			printf("Expected an expression but found %s\n", found);
		}

		if (!parseOperator(parser, &newExpr->op)) {
			const char *found = parser->tokens[parser->currentTokenIndex].repr;
			printf("Expected an operator but found %s\n", found);
		}

		if (!parseExpression(parser, &newExpr->right)) {
			const char *found = parser->tokens[parser->currentTokenIndex].repr;
			printf("Expected an expression but found %s\n", found);
		}

		if (parser->tokens[parser->currentTokenIndex].class == SEPARATOR
			&& parser->tokens[parser->currentTokenIndex].repr[0] != ')') {
			const char *found = parser->tokens[parser->currentTokenIndex].repr;
			printf("error, expected closing parenthesis, found %s\n", found);
		}

		parser->currentTokenIndex += 1;
		return true;
	}

	destroyExpression(parser, newExpr);
	return false;
}

bool parseToAST(Parser *parser, ASTNode **node) {
	Expression *expr;

	int i;
	for (i = 0; i < parser->tokenListSize; i++) {
		printf("%s\n", parser->tokens[i].repr);
	}
	printf("\n\n");

	if (parseExpression(parser, &expr)) {
		if (parser->tokens[parser->currentTokenIndex].class != END_OF_FILE) {
			printf("Shit after EOF: \n");
			printf("%s\n", parser->tokens[parser->currentTokenIndex].repr);
		}
		*node = expr;
		return true;
	}
	return 0;
}

void destroyParser(Parser *parser) {
	free(parser);
}
