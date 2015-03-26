#ifndef AST_H
#define AST_H

#include "util.h"
#include "vector.h"

/** forward declarations */
typedef struct Type s_Type;
typedef struct Expression s_Expression;
typedef struct {} ContinueStat;
typedef struct {} BreakStat;

/**
 * A list of identifiers, a common
 * production.
 */
typedef struct {
	Vector *values;
} IdentifierList;

/**
 * Different types of literals,
 * number, string, and characters.
 */
typedef enum {
	LITERAL_NUMBER,
	LITERAL_STRING,
	LITERAL_CHARACTER
} LiteralType;

/**
 * A node representing a literal.
 */
typedef struct {
	char *value;
	LiteralType type;
} Literal;

/**
 * A node representing a binary expression,
 * for example a + b is a binary expression.
 */
typedef struct {
	s_Expression *lhand;
	char operand;
	s_Expression *rhand;
} BinaryExpr;

/**
 * A node representing a unary expression, for instance
 * +5, or -5.
 */
typedef struct {
	char operand;
	s_Expression *rhand;
} UnaryExpr;

/**
 * A node representing an array sub expression,
 * for example a[5], or optional array slices, e.g a[5: 5].
 */
typedef struct {
	s_Expression *lhand;

	// optional
	s_Expression *start;
	s_Expression *end;
} ArraySubExpr;

/**
 * A node representing member access using the dot operator.
 * i.e someStruct.x
 */
typedef struct {
	s_Expression *expr;
	char *value;
} MemberAccessExpr;

/**
 * A node representing a Primary Expression, which can be either a literal,
 * a primary expression, array sub expression, member access, binary expression,
 * unary expression, etc.
 */
typedef struct {
	Literal *literal;
	s_Expression *parenExpr;
	ArraySubExpr *arraySlice;
	MemberAccessExpr *memberAccess;
} PrimaryExpr;

/**
 * An enumeration for different types of expressions,
 * this is to make it cleaner than null checking everything.
 */
typedef enum {
	PRIMARY_EXPR,
	BINARY_EXPR,
	UNARY_EXPR,
	MEMBER_ACCESS_EXPR,
	ARRAY_SLICE_EXPR,
} ExpressionType;

/**
 * Inheritance "emulation", an Expression Node is the parent
 * of all expressions.
 */
typedef struct s_Expression {
	PrimaryExpr *primary;
	BinaryExpr *binary;
	UnaryExpr *unary;
	MemberAccessExpr *memberAccess;
	ArraySubExpr *arraySlice;
} Expression;

/**
 * A name of a type.
 */
typedef struct {
	char *name;
} TypeName;

/**
 * An array type, which contains the length of the array
 * as an expression, and the type of data the array holds.
 * i.e: int[10];
 */
typedef struct {
	Expression *length;
	struct s_Type *type;
} ArrayType;

/**
 * A pointer type, i.e `int ^`
 */
typedef struct {
	struct s_Type *type;
} PointerType;

/**
 * Field declaration for a struct
 */
typedef struct {
	IdentifierList *idenList;
	struct s_Type *type;
	bool mutable;
} FieldDecl;

/**
 * List of field declarations, this could probably
 * just be in the Struct Decl as a list.
 */
typedef struct {
	Vector *members;
} FieldDeclList;

/**
 * A node representing a user-defined data structure.
 */
typedef struct {
	FieldDeclList *fields;
	char *name;
} StructDecl;

/**
 * A list of statements
 */
typedef struct {
	Vector *stmts;
} StatementList;

/**
 * Different types of blocks, again for
 * no null checking
 */
typedef enum {
	MULTI_STATEMENT_BLOCK,
	SINGLE_STATEMENT_BLOCK
} BlockType;

/**
 * A node representing a block. The grammar
 * suggests that a single statement block, denoted with
 * an arrow, only holds one statement, but we can just use
 * the same list.
 */
typedef struct {
	StatementList *stmtList;
	BlockType type;
} Block;

/**
 * This node represents a parameter in
 * a function signature
 */
typedef struct {
	s_Type *type;
	IdentifierList *identList;
	bool mutable;
} ParameterSection;

/**
 * Parameters for a function, again potentially
 * redundant, just a wrapper for a vector...
 */
typedef struct {
	Vector *paramList;
} Parameters;

/**
 * A node representing a receiver.
 */
typedef struct {
	s_Type *type;
	char *name;
	bool mutable;
} Receiver;

/**
 * A function signature or prototype.
 */
typedef struct {
	Receiver *receiver;
	char *name;
	Parameters *parameters;
	bool mutable;
} FunctionSignature;

/**
 * A function declaration, which consists
 * of a function prototype and a block.
 */
typedef struct {
	FunctionSignature *signature;
	Block *body;
} FunctionDecl;

/**
 * A node representing a variable declaration
 */
typedef struct {
	bool mutable;
	s_Type *type;
	char *name;
	s_Expression *expr;
} VariableDecl;

/**
 * A node representing all declarations,
 * function/struct/variable...
 */
typedef struct {
	FunctionDecl *funcDecl;
	StructDecl *structDecl;
	VariableDecl *varDecl;
} Declaration;

/**
 * Used for IncDecStat, if it's incrementing
 * or decrementing.
 */
typedef enum {
	// IOD = Increment or Decrement
	IOD_INCREMENT,
	IOD_DECREMENT,
} IncOrDec;

/**
 * A statement representing an increment
 * or decrement.
 */
typedef struct {
	Expression *expr;
	IncOrDec type;
} IncDecStat;

/**
 * A return statement
 */
typedef struct {
	s_Expression *expr;
} ReturnStat;

/**
 * Node for an assignment
 */
typedef struct {
	PrimaryExpr *primary;
	s_Expression *expr;
} Assignment;

/**
 * Node for all the "leave" statements,
 * e.g continue, return, break
 */
typedef struct {
	ReturnStat *retStmt;
	BreakStat *breakStmt;
	ContinueStat *conStmt;
} LeaveStat;

/**
 * Unstructured statements, i.e things that don't have
 * a body, i.e declaration, assignment...
 */
typedef struct {
	Declaration *decl;
	LeaveStat *leave;
	IncDecStat *incDec;
	Assignment *assignment;
} UnstructuredStatement;

/**
 * Else statement block
 */
typedef struct {
	Block *body;
} ElseStat;

/**
 * If statement node, expr is
 * for a condition
 */
typedef struct {
	s_Expression *expr;
	Block *body;
	ElseStat *elseStmt;
} IfStat;

/**
 * A match clause for a match statement
 */
typedef struct {
	s_Expression *expr;
	Block *body;
	LeaveStat *leave;
} MatchClause;

/**
 * match statement
 */
typedef struct {
	s_Expression *expr;
	Vector *clauses;
} MatchStat;

/**
 * node representing a for loop
 */
typedef struct {
	s_Type *type;
	char *index;
	PrimaryExpr *start, *end, *step;
	Block *body;
} ForStat;

/**
 * Structured statements, i.e for loops,
 * matches, etc..
 */
typedef struct {
	Block *block;
	IfStat *ifStmt;
	MatchStat *matchStmt;
	ForStat *forStmt;
} StructuredStatement;

/**
 * Statements, structured and unstructured
 */
typedef struct {
	StructuredStatement *structured;
	UnstructuredStatement *unstructured;
} Statement;

/**
 * A type, which can be a type name,
 * array type, pointer type, or a structure
 * declaration.
 */
typedef struct s_Type {
	TypeName *typeName;
	ArrayType *arrayType;
	PointerType *pointerType;
} Type;

#endif // AST_H
