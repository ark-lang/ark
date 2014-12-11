#ifndef PARSER_H
#define PARSER_H

#include <stdlib.h>

#include "lexer.h"
#include "util.h"
#include "vector.h"

/**
 * Parser contents
 */
typedef struct {
	Vector *tokenStream;
	Vector *parseTree;
	int tokenIndex;
	bool parsing;
} Parser;

/**
 * Different data types
 */
typedef enum {
	INTEGER, STR, DOUBLE, FLOAT, BOOL, VOID
} DataType;

/**
 * Different types of data
 * to be stored on Node Vector
 */
typedef enum {
	EXPRESSION_NODE, VARIABLE_DEF_NODE,
	VARIABLE_DEC_NODE, FUNCTION_ARG_NODE,
	FUNCTION_NODE, FUNCTION_PROT_NODE,
	BLOCK_NODE, FUNCTION_CALLEE_NODE
} NodeType;

/**
 * A wrapper for easier memory
 * management with Nodes
 */
typedef struct {
	void *data;
	NodeType type;
} Node;

/**
 * Node for an Expression
 */
typedef struct s_Expression {
	char type;
	Token *value;
	
	struct s_Expression *leftHand;
	char operand;
	struct s_Expression *rightHand;
} ExpressionNode;

/**
 * Node for an uninitialized
 * Variable
 */
typedef struct {
	DataType type;		// type of data to store
	Token *name;		// name of the variable
} VariableDefineNode;

/** 
 * Node for a Variable being declared
 */
typedef struct {
	VariableDefineNode *vdn;
	ExpressionNode *expr;
} VariableDeclareNode;

/**
 * An argument for a function
 */
typedef struct {
	DataType type;
	Token *name;
	ExpressionNode *value;
} FunctionArgumentNode;

/**
 * Function call
 */
typedef struct {
	Token *callee;
	Vector *args;
} FunctionCalleeNode;

/**
 * Node which represents a block of statements
 */
typedef struct {
	Vector *statements;
} BlockNode;

/**
 * Function prototype node
 * 
 * i.e:
 *    fn func_name(type name, type name): type
 */
typedef struct {
	Vector *args;
	Token *name;
	DataType ret;
} FunctionPrototypeNode;

/**
 * Function declaration node
 */
typedef struct {
	FunctionPrototypeNode *fpn;
	BlockNode *body;
} FunctionNode;

FunctionCalleeNode *createFunctionCalleeNode();

/**
 * Creates a new Expression Node
 * 
 * a + b
 * 1 + 2
 * (a + b) - (1 + b)
 */
ExpressionNode *createExpressionNode();

/**
 * Creates a new Variable Define node
 * 
 * int x;
 * int y;
 * double z; 
 */
VariableDefineNode *createVariableDefineNode();

/**
 * Creates a new Variable Declaration Node
 * 
 * int x = 5;
 * int d = 5 + 9;
 */
VariableDeclareNode *createVariableDeclareNode();

/**
 * Creates a new Function Argument Node
 * 
 * fn whatever(int x, int y, int z = 23): int {...
 */
FunctionArgumentNode *createFunctionArgumentNode();

/**
 * Creates a new Block Node
 * 
 * {
 *    statement;
 * }
 */
BlockNode *createBlockNode();

/**
 * Creates a new Function Node
 * 
 * fn whatever(int x, int y): int {
 *     ret x + y;
 * }
 */
FunctionNode *createFunctionNode();

/**
 * Creates a new Function Prototype Node
 * 
 * fn whatever(int x, int y): int
 */
FunctionPrototypeNode *createFunctionPrototypeNode();

void destroyFunctionCalleeNode(FunctionCalleeNode *fcn);

/**
 * Destroy an Expression Node
 */
void destroyExpressionNode(ExpressionNode *expr);

/**
 * Destroy a Variable Definition Node
 */
void destroyVariableDefineNode(VariableDefineNode *vdn);

/**
 * Destroy a Variable Declaration Node
 */
void destroyVariableDeclareNode(VariableDeclareNode *vdn);

/**
 * Destroy a Function Argument Node
 */
void destroyFunctionArgumentNode(FunctionArgumentNode *fan);

/**
 * Destroy a Block Node
 */
void destroyBlockNode(BlockNode *bn);

/**
 * Destroy a Function Prototype Node
 */
void destroyFunctionPrototypeNode(FunctionPrototypeNode *fpn);

/**
 * Destroy a Function Node
 */
void destroyFunctionNode(FunctionNode *fn);

/**
 * Prepares a node to go into a Vector, this will also
 * help with memory management
 * 
 * @param parser the parser instance for vector access
 * @param data the data to store
 * @param type the type of data
 */
void prepareNode(Parser *parser, void *data, NodeType type);

/**
 * Remove a node
 * 
 * @param node the node to remove
 */
void removeNode(Node *node);

/**
 * Create a new Parser instance
 * 
 * @param tokenStream the token stream to parse
 * @return instance of Parser
 */
Parser *parserCreate(Vector *tokenStream);

/**
 * Advances to the next token
 * 
 * @param parser parser instance
 * @return the token we consumed
 */
Token *parserConsumeToken(Parser *parser);

/**
 * Peek at the token that is {@ahead} tokens
 * away in the token stream
 * 
 * @param parser instance of parser
 * @param ahead how far ahead to peek
 * @return the Token peeking at
 */
Token *parserPeekAhead(Parser *parser, int ahead);

/**
 * Checks if the next token type is the same as the given
 * token type. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserExpectType(Parser *parser, TokenType type);

/**
 * Checks if the next tokens content is the same as the given
 * content. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserExpectContent(Parser *parser, char *content);

/**
 * Checks if the next token type is the same as the given
 * token type and the token content is the same as the given
 * content, if not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @param content content to match
 * @return the token we matched
 */
Token *parserExpectTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserMatchType(Parser *parser, TokenType type);

/**
 * Checks if the current tokens content is the same as the given
 * content. If not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *parserMatchContent(Parser *parser, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type and the token content is the same as the given
 * content, if not, throws an error
 * 
 * @param parser instance of the parser
 * @param type the type to match
 * @param content content to match
 * @return the token we matched
 */
Token *parserMatchTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * if the token at the given index is the same type as the given one
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token is the same type as the given one
 */
bool parserTokenType(Parser *parser, TokenType type, int ahead);

/**
 * if the token at the given index has the same content as the given
 * @param parser the parser instance
 * @param type the type to check
 * @param ahead how far away the token is
 * @return if the current token has the same content as the given
 */
bool parserTokenContent(Parser *parser, char* content, int ahead);

/**
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token has the same content as the given
 */
bool parserTokenTypeAndContent(Parser *parser, TokenType type, char* content, int ahead);

/**
 * Parses an expression: currently only parses a number!
 * 
 * @param parser the parser instance
 * @return the expression parsed
 */
ExpressionNode *parserParseExpression(Parser *parser);

/**
 * Prints the type and content of the current token
 */
void printCurrentToken(Parser *parser);

/**
 * Parses a variable
 * 
 * @param param the parser instance
 */
void parserParseVariable(Parser *parser);

/**
 * Parses a block of statements
 *
 * @param parser the parser instance
 */
BlockNode *parserParseBlock(Parser *parser);

/**
 * Parses a function
 * 
 * @param parser the parser instance
 */
void parserParseFunctionPrototype(Parser *parser);

/**
 * Parses a function call
 * 
 * @param parser the parser instance
 */
void parserParseFunctionCall(Parser *parser);

/**
 * Parses statements, function calls, while
 * loops, etc
 * 
 * @param parser the parser instance
 */
void parserParseStatements(Parser *parser);

/**
 * Finds the appropriate Data Type from the given Token
 * will throw an error if invalid type
 * 
 * @param parser the parser instance
 * @param tok the token to check
 * @return the token as a DataType
 */
DataType parserTokenTypeToDataType(Parser *parser, Token *tok);

/**
 * Returns if the given token is a data type
 *
 * @param parser the parser instance
 * @param tok the token instance
 * @return true if the token is a data type
 */
bool parserIsTokenDataType(Parser *parser, Token *tok);

/**
 * Start parsing
 *
 * @param parser parser to start parsing
 */
void parserStartParsing(Parser *parser);

/**
 * Destroy the given Parser
 * 
 * @param parser the parser to destroy
 */
void parserDestroy(Parser *parser);

#endif // PARSER_H