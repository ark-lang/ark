#ifndef PARSER_H
#define PARSER_H

#include <stdio.h>
#include <stdlib.h>

#include "bool.h"
#include "lexer.h"


#define TOKEN_LIST_MAX_SIZE 5000

typedef struct {
	TokenType *tokens;
	int tokenListSize;
	int currentTokenIndex;
} Parser;

typedef int Operator;

typedef enum {
	DT_INTEGER,
	DT_BOOLEAN,
	DT_VOID
} DataType;

typedef struct s_Expression {
	char type;
	int value;
	struct s_Expression *left, *right;
	Operator op;
} Expression;

typedef struct s_Variable {
	char *name;
	bool mutable;
	DataType type;
	Expression *expr;
} Variable;

typedef Expression ASTNode;

Parser *createParser(TokenType *tokens, int tokenListSize);

Expression *createExpression(Parser *parser);

void destroyExpression(Parser *parser, Expression *expr);

bool parseOperator(Parser *parser, Operator *oper);

bool parseExpression(Parser *parser, Expression **expr);

bool parseVariable(Parser *parser, Variable *var);

bool parseToAST(Parser *parser, ASTNode **node);

DataType getDataType(char *type);

void destroyParser(Parser *parser);

#endif // PARSER_H