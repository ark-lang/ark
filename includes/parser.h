#ifndef parser_H
#define parser_H

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <assert.h>

#include "lexer.h"
#include "util.h"
#include "vector.h"
#include "hashmap.h"

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
#define BLOCK_OPENER			"{"
#define BLOCK_CLOSER			"}"
#define ASSIGNMENT_OPERATOR		"="
#define FUNCTION_KEYWORD 	   	"fn"
#define VOID_KEYWORD	 	   	"void"
#define BREAK_KEYWORD	 	   	"break"
#define CONTINUE_KEYWORD		"continue"
#define RETURN_KEYWORD	 	   	"return"
#define STRUCT_KEYWORD	 	   	"struct"
#define COMMA_SEPARATOR			","
#define SEMI_COLON				";"
#define SINGLE_STATEMENT		"=>"		// cant think of a name for this operator
#define IF_KEYWORD				"if"
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
	vector *token_stream;

	// the parse tree being built
	vector *parse_tree;

	// the current token index in the stream
	int token_index;

	// if we're currently parsing
	bool parsing;

	// whether to exit on error
	// after parsing
	bool exit_on_error;

	// hashmap for validation stuff 
	hashmap *sym_table;
} parser;

/**
 * Different data types
 */
typedef enum {
	TYPE_INTEGER = 0, TYPE_STR, TYPE_DOUBLE, TYPE_FLOAT, TYPE_BOOL, TYPE_VOID,
	TYPE_CHAR, TYPE_NULL
} data_type;

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
	IF_STATEMENT_AST_NODE, MATCH_STATEMENT_AST_NODE, WHILE_LOOP_AST_NODE
} ast_node_type;

/**
 * A wrapper for easier memory
 * management with ast_nodes
 */
typedef struct {
	void *data;
	ast_node_type type;
} ast_node;

/**
 * Function call
 */
typedef struct {
	char *callee;
	vector *args;
} function_callee_ast_node;

/**
 * ast_node for an Expression
 */
typedef struct s_Expression {
	char type;
	
	token *value;

	function_callee_ast_node *function_call;

	struct s_Expression *lhand;
	char *operand;
	struct s_Expression *rhand;
} expression_ast_node;

/**
 * ast_node for an uninitialized
 * Variable
 */
typedef struct {
	data_type type;			// type of data to store
	char *name;				// name of the variable

	bool is_global;			// is it in a global scope?
	bool is_constant;		// is it a constant variable?
} variable_define_ast_node;

/**
 * ast_node for a Variable being declared
 */
typedef struct {
	variable_define_ast_node *vdn;
	expression_ast_node *expression;
} variable_declare_ast_node;

/**
 * An argument for a function
 */
typedef struct {
	data_type type;
	token *name;
	expression_ast_node *value;
} function_argument_ast_node;

/**
 * Function Return ast_node
 */
typedef struct {
	expression_ast_node *return_val;
} function_return_ast_node;

/**
 * A ast_node for containing and identifying
 * statements
 */
typedef struct {
	void *data;
	ast_node_type type;
} statement_ast_node;

/**
 * An enumeration item
 */
typedef struct {
	char *name;
	int value;
} enum_item;

/**
 * An enumeration node
 */
typedef struct {
	token *name;
	vector *enum_items;
} enumeration_ast_node;

/**
 * A node representing a break
 * from an inner loop
 */
typedef struct {
	// NOTHING! :)
} break_ast_node;

/**
 * A node representing the
 * continue keyword
 */
typedef struct {
	// yay nothing
} continue_ast_node;

/**
 * ast_node which represents a block of statements
 */
typedef struct {
	vector *statements;
} block_ast_node;

/**
 * Function prototype ast_node
 *
 * i.e:
 *    fn func_name(type name, type name): type
 */
typedef struct {
	vector *args;
	token *name;
	data_type ret;
} function_prototype_ast_node;

/**
 * Function declaration ast_node
 */
typedef struct {
	function_prototype_ast_node *fpn;
	block_ast_node *body;
	statement_ast_node *single_statement;
} function_ast_node;

/**
 * Labelled for accessing
 * certain parts of our
 * for loop
 */
typedef enum {
	FOR_START = 0,
	FOR_END,
	FOR_STEP
} for_loop_param;

/**
 * A ast_node for a for loop
 */
typedef struct {
	data_type type;			// data type
	token *index_name;		// index name
	vector *params;			// parameters (start, end, step)
	block_ast_node *body;	// contents of for loop
} for_loop_ast_node;

/**
 * ast_node for variable re-assignment
 */
typedef struct {
	token *name;
	expression_ast_node *expr;
} variable_reassignment_ast_node;

/**
 * A node for an infinite loop
 *
 * loop {
 * 		// statements
 * }
 */
typedef struct {
	block_ast_node *body;
} infinite_loop_ast_node;

/**
 * Node for a Struct
 */
typedef struct  {
	char *struct_name;
	vector *statements;
} structure_ast_node;

/**
 * ast_node to represent an if statement
 */
typedef struct {
	expression_ast_node *condition;
	block_ast_node *body;
} if_statement_ast_node;

/**
 * ast_node to represent a while loop
 */
typedef struct {
	expression_ast_node *condition;
	block_ast_node *body;
} while_ast_node;

/**
 * ast_node to represent a case for a match
 */
typedef struct {
	statement_ast_node *statement;
	block_ast_node *body;
} match_case_ast_node;

/**
 * ast_node to represent a match
 */
typedef struct {
	expression_ast_node *condition;
	vector *cases;
} match_ast_node;

/**
 * parse an operand
 */
char *parse_operand(parser *parser);

/**
 * Create a new structure node
 * @return the structure node
 */
structure_ast_node *create_structure_ast_node();

/**
 * Creates an enumeration node
 * @return the enum node we created
 */
enumeration_ast_node *create_enumeration_ast_node();

/**
 * Creates an enumeration item and fills it with values
 * @param  name  the name of the enum item
 * @param  value the value it stores
 * @return       [description]
 */
enum_item *create_enum_item(char *name, int value);

/**
 * Create an infinite loop ast node
 * @return the infinite loop ast node
 */
infinite_loop_ast_node *create_infinite_loop_ast_node();

/**
 * Creat a break ast node
 * @return the break ast node
 */
break_ast_node *create_break_ast_node();

/**
 * Create a new Variable Reassignment ast_node
 */
variable_reassignment_ast_node *create_variable_reassign_ast_node();

/**
 * Create a new For Loop ast_node
 */
for_loop_ast_node *create_for_loop_ast_node();

/**
 * Create a new  Function Callee ast_node
 */
function_callee_ast_node *create_function_callee_ast_node();

/**
 * Create a new Function Return ast_node
 */
function_return_ast_node *create_function_return_ast_node();

/**
 * Create a new Statement ast_node
 */
statement_ast_node *create_statement_ast_node();

/**
 * Creates a new Expression ast_node
 *
 * a + b
 * 1 + 2
 * (a + b) - (1 + b)
 */
expression_ast_node *create_expression_ast_node();

/**
 * Creates a new if statement ast_node
 */
if_statement_ast_node *create_if_statement_ast_node();

/**
 * Creates a new while loop ast_node
 */
while_ast_node *create_while_ast_node();

/**
 * Creates a new match case ast_node
 */
match_case_ast_node *create_match_case_ast_node();

/**
 * Creates a new match ast_node
 */
match_ast_node *create_match_ast_node();

/**
 * Creates a new Variable Define ast_node
 *
 * int x;
 * int y;
 * double z;
 */
variable_define_ast_node *create_variable_define_ast_node();

/**
 * Creates a new Variable Declaration ast_node
 *
 * int x = 5;
 * int d = 5 + 9;
 */
variable_declare_ast_node *create_variable_declare_ast_node();

/**
 * Creates a new Function Argument ast_node
 *
 * fn whatever(int x, int y, int z = 23): int {...
 */
function_argument_ast_node *create_function_argument_ast_node();

/**
 * Creates a new Block ast_node
 *
 * {
 *    statement;
 * }
 */
block_ast_node *create_block_ast_node();

/**
 * Creates a new Function ast_node
 *
 * fn whatever(int x, int y): int {
 *     ret x + y;
 * }
 */
function_ast_node *create_function_ast_node();

/**
 * Creates a new Function Prototype ast_node
 *
 * fn whatever(int x, int y): int
 */
function_prototype_ast_node *create_function_prototype_ast_node();

/**
 * Destroys the given structure
 * @param sn the node to destroy
 */
void destroy_structure_ast_node(structure_ast_node *sn);

/**
 * Destroys the given enum ast node
 * @param en the node to destroy
 */
void destroy_enumeration_ast_node(enumeration_ast_node *en);

/**
 * Destroys the given enumeration item
 * @param ei the item to destroy
 */
void destroy_enum_item(enum_item *ei);

/**
 * Destroy the break ast node
 * @param bn the node to destroy
 */
void destroy_break_ast_node(break_ast_node *bn);

/**
 * Destroy the continue ast node
 * @param bn the node to destroy
 */
void destroy_continue_ast_node(continue_ast_node *bn);


/**
 * Destroys a variable reassignement node
 * @param vrn the node to destroy
 */
void destroy_variable_reassign_ast_node(variable_reassignment_ast_node *vrn);

/**
 * Destroys an if statement node
 * @param isn the node to destroy
 */
void destroy_if_statement_ast_node(if_statement_ast_node *isn);

/**
 * Destroys a while loop node
 * @param wan the node to destroy
 */
void destroy_while_ast_node(while_ast_node *wan);

/**
 * Destroys a match case node
 * @param mcn the node to destroy
 */
void destroy_match_case_ast_node(match_case_ast_node *mcn);

/**
 * Destroys a match ast node
 * @param mn the node to destroy
 */
void destroy_match_ast_node(match_ast_node *mn);

/**
 * Destroys an infinite loop node
 * @param iln the node to destroy
 */
void destroy_infinite_loop_ast_node(infinite_loop_ast_node *iln);

/**
 * Destroys the given For Loop ast_node
 * @param fln the node to destroy
 */
void destroy_for_loop_ast_node(for_loop_ast_node *fln);

/**
 * Destroy the given Statement ast_node
 * @param sn the node to destroy
 */
void destroy_statement_ast_node(statement_ast_node *sn);

/**
 * Destroy the given Function Return ast_node
 * @param frn the node to destroy
 */
void destroy_function_return_ast_node(function_return_ast_node *frn);

/**
 * Destroy function callee ast_node
 * @param fcn the node to destroy
 */
void destroy_function_callee_ast_node(function_callee_ast_node *fcn);

/**
 * Destroy an Expression ast_node
 * @param expr the node to destroy
 */
void destroy_expression_ast_node(expression_ast_node *expr);

/**
 * Destroy a Variable Definition ast_node
 * @param vdn the node to destroy
 */
void destroy_variable_define_ast_node(variable_define_ast_node *vdn);

/**
 * Destroy a Variable Declaration ast_node
 * @param vdn the node to destroy
 */
void destroy_variable_declare_ast_node(variable_declare_ast_node *vdn);

/**
 * Destroy a Function Argument ast_node
 * @param fan the node to destroy
 */
void destroy_function_argument_ast_node(function_argument_ast_node *fan);

/**
 * Destroy a Block ast_node
 * @param bn the node to destroy
 */
void destroy_block_ast_node(block_ast_node *bn);

/**
 * Destroy a Function Prototype ast_node
 * @param fpn the node to destroy
 */
void destroy_function_prototype_ast_node(function_prototype_ast_node *fpn);

/**
 * Destroy a Function ast_node
 * @param fn the node to destroy
 */
void destroy_function_ast_node(function_ast_node *fn);

/**
 * Prepares a ast_node to go into a vector, this will also
 * help with memory management
 *
 * @param parser the parser instance for vector access
 * @param data the data to store
 * @param type the type of data
 */
void prepare_ast_node(parser *parser, void *data, ast_node_type type);

/**
 * Remove a ast_node
 *
 * @param ast_node the ast_node to remove
 */
void remove_ast_node(ast_node *ast_node);

/**
 * Create a new parser instance
 *
 * @param token_stream the token stream to parse
 * @return instance of parser
 */
parser *create_parser(vector *token_stream);

/**
 * Advances to the next token
 *
 * @param parser parser instance
 * @return the token we consumed
 */
token *consume_token(parser *parser);

/**
 * Peek at the token that is {@ahead} tokens
 * away in the token stream
 *
 * @param parser instance of parser
 * @param ahead how far ahead to peek
 * @return the token peeking at
 */
token *peek_at_token_stream(parser *parser, int ahead);

/**
 * Checks if the next token type is the same as the given
 * token type. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
token *expect_token_type(parser *parser, token_type type);

/**
 * Checks if the next tokens content is the same as the given
 * content. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
token *expect_token_content(parser *parser, char *content);

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
token *expect_token_type_and_content(parser *parser, token_type type, char *content);

/**
 * Checks if the current token type is the same as the given
 * token type. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
token *match_token_type(parser *parser, token_type type);

/**
 * Checks if the current tokens content is the same as the given
 * content. If not, throws an error
 *
 * @param parser instance of the parser
 * @param type the type to match
 * @return the token we matched
 */
token *match_token_content(parser *parser, char *content);

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
token *match_token_type_and_content(parser *parser, token_type type, char *content);

/**
 * if the token at the given index is the same type as the given one
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token is the same type as the given one
 */
bool check_token_type(parser *parser, token_type type, int ahead);

/**
 * if the token at the given index has the same content as the given
 * @param parser the parser instance
 * @param type the type to check
 * @param ahead how far away the token is
 * @return if the current token has the same content as the given
 */
bool check_token_content(parser *parser, char* content, int ahead);

/**
 * @param parser the parser instance
 * @param type the type to check
 * @return if the current token has the same content as the given
 */
bool check_token_type_and_content(parser *parser, token_type type, char* content, int ahead);

/**
 * Parses an expression: currently only parses a number!
 *
 * @param parser the parser instance
 * @return the expression parsed
 */
expression_ast_node *parse_expression_ast_node(parser *parser);

/**
 * Parses a For Loop statement
 */
statement_ast_node *parse_for_loop_ast_node(parser *parser);

/**
 * Prints the type and content of the current token
 */
void print_current_token(parser *parser);

/**
 * Parses a variable
 *
 * @param param the parser instance
 * @param global if the variable is globally declared
 */
void *parse_variable_ast_node(parser *parser, bool global);

/**
 * Parses a block of statements
 *
 * @param parser the parser instance
 */
block_ast_node *parse_block_ast_node(parser *parser);

/**
 * Parses an infinite loop ast node
 * @param  parser the parser to parse with
 * @return        the loop node as a statement node
 */
statement_ast_node *parse_infinite_loop_ast_node(parser *parser);

/**
 * Parses a function
 *
 * @param parser the parser instance
 */
function_ast_node *parse_function_ast_node(parser *parser);

/**
 * Parses a function call
 *
 * @param parser the parser instance
 */
function_callee_ast_node *parse_function_callee_ast_node(parser *parser);

/**
 * Parses statements, function calls, while
 * loops, etc
 *
 * @param parser the parser instance
 */
statement_ast_node *parse_statement_ast_node(parser *parser);

/**
 * Finds the appropriate Data Type from the given token
 * will throw an error if invalid type
 *
 * @param parser the parser instance
 * @param tok the token to check
 * @return the token as a data_type
 */
data_type match_token_type_to_data_type(parser *parser, token *tok);

/**
 * Returns if the given token is a data type
 *
 * @param parser the parser instance
 * @param tok the token instance
 * @return true if the token is a data type
 */
bool check_token_type_is_valid_data_type(parser *parser, token *tok);

/**
 * Parses a return statement node
 * @param  parser the parser to parse with
 * @return        the return ast node
 */
function_return_ast_node *parse_return_statement_ast_node(parser *parser);

/**
 * Parses a structure node
 * @param  parser the parser the parse with
 * @return        the sturct node
 */
structure_ast_node *parse_structure_ast_node(parser *parser);

/**
 * Parses a variable reassignment
 *
 * @parser the parser instance
 */
variable_reassignment_ast_node *parse_reassignment_statement_ast_node(parser *parser);

/**
 * Start parsing
 *
 * @param parser parser to start parsing
 */
void start_parsing_token_stream(parser *parser);

/**
 * Destroy the given parser
 *
 * @param parser the parser to destroy
 */
void destroy_parser(parser *parser);

#endif // parser_H
