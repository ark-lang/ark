#ifndef PARSER_H
#define PARSER_H

#include <stdio.h>
#include <stdlib.h>

#include "lexer.h"

#define TOKEN_LIST_MAX_SIZE 5000
#define false 0
#define true 1

typedef struct {
	TokenType *tokens;
	int tokenListSize;
	int currentTokenIndex;
} Parser;

typedef int bool;
typedef int Operator;

typedef struct s_Expression {
	char type;
	int value;
	struct s_Expression *left, *right;
	Operator op;
} Expression;

typedef Expression ASTNode;

Parser *createParser(TokenType *tokens, int tokenListSize);

Expression *createExpression(Parser *parser);

void destroyExpression(Parser *parser, Expression *expr);

int parseOperator(Parser *parser, Operator *oper);

int parseExpression(Parser *parser, Expression **expr);

bool parseToAST(Parser *parser, ASTNode **node);

void destroyParser(Parser *parser);

#endif // PARSER_H