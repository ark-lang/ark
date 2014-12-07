#ifndef PARSER_H
#define PARSER_H

typedef struct {
	double value;
} NumberExpr;

typedef struct {
	char* name;
} VariableExpr;

typedef struct {
	char operand;
	Expr *lhs;
	Expr *rhs;
} BinaryExpr;

typedef struct {
	Vector *tokenStream;
} Parser;

Parser *parserCreate(Vector *tokenStream);

void parserDestroy(Parser *parser);

#endif // PARSER_H