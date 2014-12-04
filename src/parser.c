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

bool parseVariable(Parser *parser, Variable *var) {
	// check if variable is mutable
	bool mutable = false;
	if (!strcmp(parser->tokens[parser->currentTokenIndex].repr, "mut")) {
		mutable = true;
		parser->currentTokenIndex += 1;
	}
	
	// check for data type
	if (parser->tokens[parser->currentTokenIndex].class == IDENTIFIER) {
		if (isDataType(parser->tokens[parser->currentTokenIndex].repr)) {
			// get data type as enum
			DataType type = getDataType(parser->tokens[parser->currentTokenIndex].repr);
			parser->currentTokenIndex += 1;

			// check for name of variable
			if (parser->tokens[parser->currentTokenIndex].class == IDENTIFIER) {
				// store
				char *variableName = parser->tokens[parser->currentTokenIndex].repr;
				parser->currentTokenIndex += 1;

				if (parser->tokens[parser->currentTokenIndex].class == OPERATOR) {
					Expression *variableExpr = createExpression(parser);

					if (parser->tokens[parser->currentTokenIndex].repr[0] == '=') {
						parser->currentTokenIndex += 1;

						if (!parseExpression(parser, &variableExpr)) {
							char *found = parser->tokens[parser->currentTokenIndex].repr;
							printf("error: expected variable assignment expression, found %s\n", found);
							exit(1);
						}
						else {
							var->name = variableName;
							var->mutable = mutable;
							var->type = type;
							var->expr = variableExpr;
							return true;
						}
					}
					else if (parser->tokens[parser->currentTokenIndex].repr[0] == ';') {
						variableExpr->type = 'D';
						variableExpr->value = 0;

						var->name = variableName;
						var->mutable = mutable;
						var->type = type;
						var->expr = variableExpr;
						return true;
					}
					else {
						char *found = parser->tokens[parser->currentTokenIndex].repr;
						printf("error: expected operator, found %s\n", found);
						exit(1);
					}
				}
			}
			else {
				char *found = parser->tokens[parser->currentTokenIndex].repr;
				printf("error: expected variable name, found %s\n", found);
				exit(1);
			}
		}
		else {
			char *found = parser->tokens[parser->currentTokenIndex].repr;
			printf("error: expected data type, found %s\n", found);
			exit(1);
		}
	}
	return false;
}

DataType getDataType(char *type) {
	if (!strcmp(type, "int")) {
		return DT_INTEGER;
	}
	else if (!strcmp(type, "bool")) {
		return DT_BOOLEAN;
	}
	return DT_VOID;
}

bool parseToAST(Parser *parser, ASTNode **node) {
	Variable *var;

	if (parseVariable(parser, var)) {
		if (parser->tokens[parser->currentTokenIndex].class != END_OF_FILE) {
			printf("Shit after EOF: \n");
			printf("%s\n", parser->tokens[parser->currentTokenIndex].repr);
		}
		return true;
	}
	return false;
}

void destroyParser(Parser *parser) {
	free(parser);
}
