#ifndef PARSER_H
#define PARSER_H

/**
 * Node for a Number
 */
typedef struct {
	double value;
} NumberExpr;

/**
 * Node for a variable call
 * e.g int a = b + c;
 */
typedef struct {
	char* name;
} VariableExpr;

/**
 * Node for a binary expression,
 * e.g
 * a + 5
 */
typedef struct {
	char operand;
	Expr *lhs;
	Expr *rhs;
} BinaryExpr;

typedef struct {
	Vector *tokenStream;
} Parser;

/**
 * Create a new Parser instance
 * 
 * @param tokenStream the token stream to parse
 * @return instance of Parser
 */
Parser *parserCreate(Vector *tokenStream);

/**
 * Destroy the given Parser
 * 
 * @param parser the parser to destroy
 */
void parserDestroy(Parser *parser);

#endif // PARSER_H