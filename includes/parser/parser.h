#ifndef parser_H
#define parser_H

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <assert.h>

#include "lexer/lexer.h"
#include "util/util.h"
#include "util/vector.h"
#include "util/hashmap.h"

#define MAX_FOR_LOOP_PARAM_COUNT 3
#define MIN_FOR_LOOP_PARAM_COUNT 2

// for defining expression types
// values are somewhat arbitrary
#define EXPR_LOGICAL_OPERATOR 	'L'
#define EXPR_NUMBER				'N'
#define EXPR_STRING 			'S'
#define EXPR_CHARACTER 			'C'
#define EXPR_VARIABLE 			'V'
#define EXPR_PARENTHESIS 		'P'
#define EXPR_FUNCTION_CALL		'F'

// keywords
#define CONSTANT_KEYWORD 	   	"const"
#define ANON_STRUCT_KEYWORD		"anon"
#define BLOCK_OPENER			"{"
#define BLOCK_CLOSER			"}"
#define ASSIGNMENT_OPERATOR		"="
#define POINTER_OPERATOR		"^"
#define ADDRESS_OF_OPERATOR		"&"
#define FUNCTION_KEYWORD 	   	"fn"
#define VOID_KEYWORD	 	   	"void"
#define BREAK_KEYWORD	 	   	"break"
#define TRUE_KEYWORD			"true"
#define FALSE_KEYWORD			"false"
#define CONTINUE_KEYWORD		"continue"
#define RETURN_KEYWORD	 	   	"return"
#define STRUCT_KEYWORD	 	   	"struct"
#define COMMA_SEPARATOR			","
#define SEMI_COLON				";"
#define SINGLE_STATEMENT		"->"		// cant think of a name for this operator
#define IF_KEYWORD				"if"
#define ELSE_KEYWORD			"else"
#define MATCH_KEYWORD			"match"
#define ENUM_KEYWORD	 	   	"enum"
#define UNSAFE_KEYWORD	 	   	"unsafe"
#define UNDERSCORE_KEYWORD		"_"			// underscores are treated as identifiers
#define IF_STATEMENT_KEYWORD   	"if"
#define WHILE_LOOP_KEYWORD	   	"while"
#define INFINITE_LOOP_KEYWORD  	"loop"
#define ELSE_KEYWORD		   	"else"
#define MATCH_KEYWORD			"match"
#define FOR_LOOP_KEYWORD		"for"
#define TRUE_KEYWORD			"true"
#define FALSE_KEYWORD			"false"

/**
 * parser contents
 */
typedef struct {
	// the token stream to parse
	Vector *tokenStream;

	// the parse tree being built
	Vector *parseTree;

	// the current token index in the stream
	int tokenIndex;

	// if we're currently parsing
	bool parsing;

	// whether to exit on error
	// after parsing
	bool exitOnError;

	// timers for benchmarking
	clock_t timer;
	double secondsTaken;
	double msTaken;
} Parser;

/**
 * Different data types
 */
typedef enum {
	TYPE_INTEGER = 0, TYPE_STR, TYPE_DOUBLE, TYPE_FLOAT, TYPE_BOOL, TYPE_VOID,
	TYPE_CHAR, TYPE_STRUCT, TYPE_NULL
} DataType;

/**
 * Different types of data
 * to be stored on ast_node vector
 */
typedef enum {
	EXPRESSION_AST_NODE = 0, VARIABLE_DEF_AST_NODE,
	VARIABLE_DEC_AST_NODE, FUNCTION_ARG_AST_NODE,
	FUNCTION_AST_NODE, FUNCTION_PROT_AST_NODE,
	BLOCK_AST_NODE, FUNCTION_CALLEE_AST_NODE,
	FUNCTION_RET_AST_NODE, FOR_LOOP_AST_NODE,
	VARIABLE_REASSIGN_AST_NODE, INFINITE_LOOP_AST_NODE,
	BREAK_AST_NODE, CONTINUE_AST_NODE, ENUM_AST_NODE, STRUCT_AST_NODE,
	IF_STATEMENT_AST_NODE, MATCH_STATEMENT_AST_NODE, WHILE_LOOP_AST_NODE,
	ANON_AST_NODE
} AstNodeType;

/**
 * A wrapper for easier memory
 * management with ast_nodes
 */
typedef struct {
	void *data;
	AstNodeType type;
} AstNode;

/**
 * Node for a Struct
 */
typedef struct  {
	char *name;
	Vector *statements;
} StructureAstNode;

/**
 * Function call
 */
typedef struct {
	char *name;
	Vector *args;
} FunctionCallAstNode;

/**
 * pointer types, e.g are we
 * dereferencing, getting the address of something,
 * or is it unspecified
 */
typedef enum {
	DEREFERENCE,
	ADDRESS_OF,
	UNSPECIFIED
} ExpressionPointerOperation;

/**
 * ast_node for an Expression
 */
typedef struct {
	Vector *expressionValues;
} ExpressionAstNode;

/**
 * ast_node for an uninitialized
 * Variable
 */
typedef struct {
	Token *type;
	char *name;				// name of the variable
	StructureAstNode *owner; // the owner of the variable?

	bool isGlobal;			// is it in a global scope?
	bool isConstant;		// is it a constant variable?
	bool isPointer;		// is it a pointer?
} VariableDefinitionAstNode;

/**
 * ast_node for a Variable being declared
 */
typedef struct {
	VariableDefinitionAstNode *variableDefinitionAstNode;
	ExpressionAstNode *expression;
} VariableDeclarationAstNode;

/**
 * The owner of a function, i.e what makes
 * a function a method
 */
typedef struct {
	Token *owner;
	Token *alias;

	bool isPointer;
} FunctionOwnerAstNode;

/**
 * An argument for a function
 */
typedef struct {
	Token *type;
	Token *name;
	ExpressionAstNode *value;

	bool isPointer;
	bool isConstant;
} FunctionArgumentAstNode;

/**
 * Function Return ast_node
 */
typedef struct {
	ExpressionAstNode *returnValue;
} FunctionReturnAstNode;

/**
 * A ast_node for containing and identifying
 * statements
 */
typedef struct {
	void *data;
	AstNodeType type;
} StatementAstNode;

/**
 * An enumeration item
 */
typedef struct {
	char *name;
	int value;
} EnumItem;

/**
 * An enumeration node
 */
typedef struct {
	Token *name;
	Vector *enumItems;
} EnumAstNode;

/**
 * A node representing a break
 * from an inner loop
 */
typedef struct {
	// n/a
} BreakStatementAstNode;

/**
 * A node representing the
 * continue keyword
 */
typedef struct {
	// n/a
} ContinueStatementAstNode;

/**
 * ast_node which represents a block of statements
 */
typedef struct {
	Vector *statements;
	bool isSingleStatement;
} BlockAstNode;

/**
 * Function prototype ast_node
 *
 * i.e:
 *    fn func_name(type name, type name): type
 */
typedef struct {
	/** function arguments */
	Vector *args;

	/** function owner */
	FunctionOwnerAstNode *owner;
	
	/** name of the function */
	Token *name;
	
	/**
	 * the return type of the function, this is a token
	 * as the return type can be a structure, not just a data type!
	 */
	Token *returnType;
} FunctionPrototypeAstNode;

/**
 * Function declaration ast_node
 */
typedef struct {
	/** function prototype of the function */
	FunctionPrototypeAstNode *prototype;

	/** functions body */
	BlockAstNode *body;

	/** if the function is a single statement, i.e -> */
	StatementAstNode *isSingleStatement;

	/** does the function return a pointer */
	bool returnsPointer;

	/** does the function return a constant value */
	bool isConstant;
} FunctionAstNode;

/**
 * Labelled for accessing
 * certain parts of our
 * for loop
 */
typedef enum {
	FOR_START = 0,
	FOR_END,
	FOR_STEP
} ForLoopParameters;

/**
 * A ast_node for a for loop
 */
typedef struct {
	/** the for loop type */
	Token *type;

	/** the name of the for loops index i.e current iteration variable */
	Token *indexName;		

	/** parameters for the for loop */
	Vector *parameters;			

	/** for loops body, things to do every iteration */
	BlockAstNode *body;	
} ForLoopAstNode;

/**
 * ast_node for variable re-assignment
 */
typedef struct {
	/** the variable being re-assigned, or it's identifier */
	Token *name;

	/** the expression to re-assign it to */
	ExpressionAstNode *expression;
} VariableReassignmentAstNode;

/**
 * an abstract syntax node for an enumerated structure,
 * it contains a name which is stored as a token so we can
 * get additional information for error checking. It also contains
 * a vector, which will store all of the structs that we're linking
 * too (as Tokens)
 */
typedef struct {
	Token *name;
	Vector *structs;
} EnumeratedStructureAstNode;

/**
 * A node for an infinite loop
 */
typedef struct {
	/** things to do while looping */
	BlockAstNode *body;
} InfiniteLoopAstNode;

/**
 * Helper enumeration for the if
 * statements
 */
typedef enum {
	IF_STATEMENT,
	ELSE_STATEMENT
} STATEMENT_TYPE;

/**
 * ast_node to represent an if statement
 */
typedef struct {
	/** statement type */
	int statementType;

	/** the condition to check */
	ExpressionAstNode *condition;

	/** body of the if statement */
	BlockAstNode *body;

	/** things to do if the condition is false */
	BlockAstNode *elseStatement;
} IfStatementAstNode;

/**
 * ast_node to represent a while loop
 */
typedef struct {
	/** condition to check */
	ExpressionAstNode *condition;

	/** things to do during each iteration */
	BlockAstNode *body;
} WhileLoopAstNode;

/**
 * ast_node to represent a case for a match
 */
typedef struct {
	/** the condition of the match */
	ExpressionAstNode *condition;

	/** things to do if the cases condition is true */
	BlockAstNode *body;
} MatchCaseAstNode;

/**
 * ast_node to represent a match
 */
typedef struct {
	/** what value we are matching */
	ExpressionAstNode *condition;

	/** the cases in the match statement */
	Vector *cases;
} MatchAstNode;

/**
 * Attempts to safely exit from the parser
 * @param parser the parser to exit
 */
void exitParser(Parser *parser);

/**
 * parse an operand
 */
int parseOperand(Parser *parser);

/**
 * Create an enumerated structure abstract syntax tree node
 * @return the ast node we created
 */
EnumeratedStructureAstNode *createEnumeratedStructureAstNode();

/**
 * Create a new structure node
 * @return the structure node
 */
StructureAstNode *createStructureAstNode();

/**
 * Creates an enumeration node
 * @return the enum node we created
 */
EnumAstNode *createEnumerationAstNode();

/**
 * Creates an enumeration item and fills it with values
 * @param  name  the name of the enum item
 * @param  value the value it stores
 * @return       [description]
 */
EnumItem *createEnumItem(char *name, int value);

/**
 * Creats a function owner ast node
 * @return the function owner struct address thing
 */
FunctionOwnerAstNode *createFunctionOwnerAstNode();

/**
 * Create an infinite loop ast node
 * @return the infinite loop ast node
 */
InfiniteLoopAstNode *createInfiniteLoopAstNode();

/**
 * Creat a break ast node
 * @return the break ast node
 */
BreakStatementAstNode *createBreakStatementAstNode();

/**
 * Create a new Variable Reassignment ast_node
 */
VariableReassignmentAstNode *createVariableReassignmentAstNode();

/**
 * Create a new For Loop ast_node
 */
ForLoopAstNode *createForLoopAstNode();

/**
 * Create a new  Function Call ast_node
 */
FunctionCallAstNode *createFunctionCallAstNode();

/**
 * Create a new Function Return ast_node
 */
FunctionReturnAstNode *createFunctionReturnAstNode();

/**
 * Create a new Statement ast_node
 */
StatementAstNode *createStatementAstNode();

/**
 * Creates a new Expression ast_node
 *
 * a + b
 * 1 + 2
 * (a + b) - (1 + b)
 */
ExpressionAstNode *createExpressionAstNode();

/**
 * Creates a new if statement ast_node
 */
IfStatementAstNode *createIfStatementAstNode();

/**
 * Creates a new while loop ast_node
 */
WhileLoopAstNode *createWhileLoopAstNode();

/**
 * Creates a new match case ast_node
 */
MatchCaseAstNode *createMatchCaseAstNode();

/**
 * Creates a new match ast_node
 */
MatchAstNode *createMatchAstNode();

/**
 * Creates a new Variable Define ast_node
 *
 * int x;
 * int y;
 * double z;
 */
VariableDefinitionAstNode *createVariableDefinitionAstNode();

/**
 * Creates a new Variable Declaration ast_node
 *
 * int x = 5;
 * int d = 5 + 9;
 */
VariableDeclarationAstNode *createVariableDeclarationAstNode();

/**
 * Creates a new Function Argument ast_node
 *
 * fn whatever(int x, int y, int z = 23): int {...
 */
FunctionArgumentAstNode *createFunctionArgumentAstNode();

/**
 * Creates a new Block ast_node
 *
 * {
 *    statement;
 * }
 */
BlockAstNode *createBlockAstNode();

/**
 * Creates a new Function ast_node
 *
 * fn whatever(int x, int y): int {
 *     ret x + y;
 * }
 */
FunctionAstNode *createFunctionAstNode();

/**
 * Creates a new Function Prototype ast_node
 *
 * fn whatever(int x, int y): int
 */
FunctionPrototypeAstNode *createFunctionPrototypeAstNode();

/**
 * Destroys the given enumerated structure ast node
 * @param es es the node to destroy
 */
void destroyEnumeratedStructureAstNode(EnumeratedStructureAstNode *es);

/**
 * Destroys the given structure
 * @param sn the node to destroy
 */
void destroyStructureAstNode(StructureAstNode *sn);

/**
 * Destroys the given enum ast node
 * @param en the node to destroy
 */
void destroyEnumAstNode(EnumAstNode *en);

/**
 * Destroys the given enumeration item
 * @param ei the item to destroy
 */
void destroyEnumItem(EnumItem *ei);

/**
 * Destroys the given function owner ast node
 * @param fo the function owner ast node to destroy
 */
void destroyFunctionOwnerAstNode(FunctionOwnerAstNode *fo);

/**
 * Destroy the break ast node
 * @param bn the node to destroy
 */
void destroyBreakStatementAstNode(BreakStatementAstNode *bn);

/**
 * Destroy the continue ast node
 * @param bn the node to destroy
 */
void destroyContinueStatementAstNode(ContinueStatementAstNode *bn);

/**
 * Destroys a variable reassignement node
 * @param vrn the node to destroy
 */
void destroyVariableReassignmentAstNode(VariableReassignmentAstNode *vrn);

/**
 * Destroys an if statement node
 * @param isn the node to destroy
 */
void destroyIfStatementAstNode(IfStatementAstNode *isn);

/**
 * Destroys a while loop node
 * @param wan the node to destroy
 */
void destroyWhileLoopAstNode(WhileLoopAstNode *wan);

/**
 * Destroys a match case node
 * @param mcn the node to destroy
 */
void destroyMastCaseAstNode(MatchCaseAstNode *mcn);

/**
 * Destroys a match ast node
 * @param mn the node to destroy
 */
void destroyMatchAstNode(MatchAstNode *mn);

/**
 * Destroys an infinite loop node
 * @param iln the node to destroy
 */
void destroyInfiniteLoopAstNode(InfiniteLoopAstNode *iln);

/**
 * Destroys the given For Loop ast_node
 * @param fln the node to destroy
 */
void destroyForLoopAstNode(ForLoopAstNode *fln);

/**
 * Destroy the given Statement ast_node
 * @param sn the node to destroy
 */
void destroyStatementAstNode(StatementAstNode *sn);

/**
 * Destroy the given Function Return ast_node
 * @param frn the node to destroy
 */
void destroyFunctionReturnAstNode(FunctionReturnAstNode *frn);

/**
 * Destroy function callee ast_node
 * @param fcn the node to destroy
 */
void destroyFunctionCallAstNode(FunctionCallAstNode *fcn);

/**
 * Destroy an Expression ast_node
 * @param expr the node to destroy
 */
void destroyExpressionAstNode(ExpressionAstNode *expr);

/**
 * Destroy a Variable Definition ast_node
 * @param vdn the node to destroy
 */
void destroyVariableDefinitionAstNode(VariableDefinitionAstNode *vdn);

/**
 * Destroy a Variable Declaration ast_node
 * @param vdn the node to destroy
 */
void destroyVariableDeclarationAstNode(VariableDeclarationAstNode *vdn);

/**
 * Destroy a Function Argument ast_node
 * @param fan the node to destroy
 */
void destroyFunctionArgumentAstNode(FunctionArgumentAstNode *fan);

/**
 * Destroy a Block ast_node
 * @param bn the node to destroy
 */
void destroyBlockAstNode(BlockAstNode *bn);

/**
 * Destroy a Function Prototype ast_node
 * @param fpn the node to destroy
 */
void destroyFunctionPrototypeAstNode(FunctionPrototypeAstNode *fpn);

/**
 * Destroy a Function ast_node
 * @param fn the node to destroy
 */
void destroyFunctionAstNode(FunctionAstNode *fn);

/**
 * Prepares a ast_node to go into a vector, this will also
 * help with memory management
 *
 * @param parser the parser instance for vector access
 * @param data the data to store
 * @param type the type of data
 */
void pushAstNode(Parser *parser, void *data, AstNodeType type);

/**
 * Remove a ast_node
 *
 * @param ast_node the ast_node to remove
 */
void removeAstNode(AstNode *astNode);

/**
 * Create a new parser instance
 *
 * @param token_stream the token stream to parse
 * @return instance of parser
 */
Parser *createParser(Vector *tokenStream);

/**
 * Advances to the next token
 *
 * @param parser parser instance
 * @return the token we consumed
 */
Token *consumeToken(Parser *parser);

/**
 * Peek at the token that is {@ahead} tokens
 * away in the token stream
 *
 * @param parser instance of parser
 * @param ahead how far ahead to peek
 * @return the token peeking at
 */
Token *peekAtTokenStream(Parser *parser, int ahead);

/**
 * Checks if the next token type is the same as the given
 * token type. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *expectTokenType(Parser *parser, TokenType type);

/**
 * Checks if the next tokens content is the same as the given
 * content. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *expectTokenContent(Parser *parser, char *content);

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
Token *expectTokenTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *matchTokenType(Parser *parser, TokenType type);

/**
 * Checks if the current tokens content is the same as the given
 * content. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
Token *matchTokenContent(Parser *parser, char *content);

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
Token *matchTokenTypeAndContent(Parser *parser, TokenType type, char *content);

/**
 * if the token at the given index is the same type as the given one
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token is the same type as the given one
 */
bool checkTokenType(Parser *parser, TokenType type, int ahead);

/**
 * if the token at the given index has the same content as the given
 * @param parser the parser instance
 * @param type the type to check
 * @param ahead how far away the token is
 * @return if the current token has the same content as the given
 */
bool checkTokenContent(Parser *parser, char* content, int ahead);

/**
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token has the same content as the given
 */
bool checkTokenTypeAndContent(Parser *parser, TokenType type, char* content, int ahead);

/**
 * Parses an expression: currently only parses a number!
 *
 * @param parser the parser instance
 * @return the expression parsed
 */
ExpressionAstNode *parseExpressionAstNode(Parser *parser);

/**
 * Parses a For Loop statement
 */
StatementAstNode *parseForLoopAstNode(Parser *parser);

/**
 * Parses a variable
 *
 * @param param the parser instance
 * @param global if the variable is globally declared
 */
void *parseVariableAstNode(Parser *parser, bool global);

/**
 * Parses a block of statements
 *
 * @param parser the parser instance
 */
BlockAstNode *parseBlockAstNode(Parser *parser);

/**
 * Parses an infinite loop ast node
 * @param  parser the parser to parse with
 * @return        the loop node as a statement node
 */
StatementAstNode *parseInfiniteLoopAstNode(Parser *parser);

/**
 * Parses a function
 *
 * @param parser the parser instance
 */
FunctionAstNode *parseFunctionAstNode(Parser *parser);

/**
 * Parses a function call
 *
 * @param parser the parser instance
 */
FunctionCallAstNode *parseFunctionCallAstNode(Parser *parser);

/**
 * Parses statements, function calls, while
 * loops, etc
 *
 * @param parser the parser instance
 */
StatementAstNode *parseStatementAstNode(Parser *parser);

/**
 * Returns if the given token is a data type
 *
 * @param parser the parser instance
 * @param tok the token instance
 * @return true if the token is a data type
 */
bool validateTokenType(Parser *parser, Token *tok);

/**
 * Parses a return statement node
 * @param  parser the parser to parse with
 * @return        the return ast node
 */
FunctionReturnAstNode *parseReturnStatementAstNode(Parser *parser);

/**
 * Parses a structure node
 * @param  parser the parser the parse with
 * @return        the sturct node
 */
StructureAstNode *parseStructureAstNode(Parser *parser);

/**
 * Parses a variable reassignment
 *
 * @parser the parser instance
 */
VariableReassignmentAstNode *parseReassignmentAstNode(Parser *parser);

/**
 * Start parsing
 *
 * @param parser parser to start parsing
 */
void parseTokenStream(Parser *parser);

/**
 * Destroy the given parser
 *
 * @param parser the parser to destroy
 */
void destroyParser(Parser *parser);

#endif // parser_H
