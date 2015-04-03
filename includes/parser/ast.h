#ifndef AST_H
#define AST_H

#include "util.h"
#include "vector.h"

/** forward declarations */
typedef struct s_Expression Expression;
typedef struct s_Type Type;
typedef struct {} ContinueStat;
typedef struct {} BreakStat;

/** different types of node for a cleanup */
typedef enum {
	IDENTIFIER_LIST_NODE, IDENTIFIER_NODE, LITERAL_NODE, BINARY_EXPR_NODE,
	UNARY_EXPR_NODE, ARRAY_SUB_EXPR_NODE, MEMBER_ACCESS_NODE,
	PRIMARY_EXPR_NODE, EXPR_NODE, TYPE_NAME_NODE, TYPE_LIT_NODE, PAREN_EXPR_NODE,
	ARRAY_TYPE_NODE, POINTER_TYPE_NODE, FIELD_DECL_NODE,
	FIELD_DECL_LIST_NODE, STRUCT_DECL_NODE, STATEMENT_LIST_NODE,
	BLOCK_NODE, PARAMETER_SECTION_NODE, PARAMETERS_NODE, RECEIVER_NODE,
	FUNCTION_SIGNATURE_NODE, FUNCTION_DECL_NODE, VARIABLE_DECL_NODE, FUNCTION_CALL_NODE,
	DECLARATION_NODE, INC_DEC_STAT_NODE, RETURN_STAT_NODE, BREAK_STAT_NODE,
	CONTINUE_STAT_NODE, LEAVE_STAT_NODE, ASSIGNMENT_NODE, UNSTRUCTURED_STATEMENT_NODE,
	ELSE_STAT_NODE, IF_STAT_NODE, MATCH_CLAUSE_STAT, MATCH_STAT_NODE, FOR_STAT_NODE,
	STRUCTURED_STATEMENT_NODE, STATEMENT_NODE, TYPE_NODE
} NodeType;

/**
 * Wrapper for a node
 */
typedef struct {
	void *data;
	NodeType type;
} Node;

/**
 * A list of identifiers, a common
 * production.
 */
typedef struct {
	Vector *values;
} IdentifierList;

/**
 * A node representing a literal.
 */
typedef struct {
	char *value;
	int type;
} Literal;

/**
 * A name of a type.
 */
typedef struct {
	char *name;
} TypeName;

typedef struct {
	TypeName *type;
} BaseType;

/**
 * A pointer type, i.e `int ^`
 */
typedef struct {
	BaseType *baseType;
} PointerType;

/**
 * An array type, which contains the length of the array
 * as an expression, and the type of data the array holds.
 * i.e: int[10];
 */
typedef struct {
	Expression *length;
	Type *type;
} ArrayType;

typedef struct {
	ArrayType *arrayType;
	PointerType *pointerType;
	int type;
} TypeLit;

/**
 * A type, which can be a type name,
 * array type, pointer type, or a structure
 * declaration.
 */
struct s_Type {
	TypeName *typeName;
	TypeLit *typeLit;
	int type;
};

/**
 * A node representing an array sub expression,
 * for example a[5], or optional array slices, e.g a[5: 5].
 */
typedef struct {
	Expression *lhand;
	Expression *start;
	Expression *end;
} ArraySubExpr;

/**
 * A node representing member access using the dot operator.
 * i.e someStruct.x
 */
typedef struct {
	Expression *expr;
	char *value;
} MemberAccessExpr;

/**
 * A node representing a function call
 */
typedef struct {
	Expression *callee;
	Vector *arguments;
} Call;

/**
 * A node representing a unary expression, for instance
 * +5, or -5.
 */
typedef struct s_UnaryExpr {
	Expression *lhand;
	char *unaryOp;
} UnaryExpr;

typedef struct {
	Expression *lhand;
	char *binaryOp;
	Expression *rhand;
} BinaryExpr;

/**
 * Inheritance "emulation", an Expression Node is the parent
 * of all expressions.
 */
struct s_Expression {
	Call *call;
	Type *type;
	Literal *lit;
	BinaryExpr *binary;
	UnaryExpr *unary;
	int exprType;
};

/**
 * Field declaration for a struct
 */
typedef struct {
	IdentifierList *idenList;
	Type *type;
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
 * A node representing a block. The grammar
 * suggests that a single statement block, denoted with
 * an arrow, only holds one statement, but we can just use
 * the same list.
 */
typedef struct {
	StatementList *stmtList;
} Block;

/**
 * This node represents a parameter in
 * a function signature
 */
typedef struct {
	Type *type;
	char *name;
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
	Type *type;
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
	Type *type;
} FunctionSignature;

/**
 * A function, which consists
 * of a function prototype and an optional block.
 */
typedef struct {
	FunctionSignature *signature;
	Block *body;
	bool prototype;
} FunctionDecl;

/**
 * A node representing a variable declaration
 */
typedef struct {
	bool mutable;
	Type *type;
	char *name;
	Expression *expr;
} VariableDecl;

/**
 * A node representing all declarations,
 * function/struct/variable...
 */
typedef struct {
	FunctionDecl *funcDecl;
	StructDecl *structDecl;
	VariableDecl *varDecl;
	int type;
} Declaration;

/**
 * A statement representing an increment
 * or decrement.
 */
typedef struct {
	Expression *expr;
	int amount;
} IncDecStat;

/**
 * A return statement
 */
typedef struct {
	Expression *expr;
} ReturnStat;

/**
 * Node for an assignment
 */
typedef struct {
	Expression *primary;
	Expression *expr;
} Assignment;

/**
 * Node for all the "leave" statements,
 * e.g continue, return, break
 */
typedef struct {
	ReturnStat *retStmt;
	BreakStat *breakStmt;
	ContinueStat *conStmt;
	int type;
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
	int type;
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
	Expression *expr;
	Block *body;
	ElseStat *elseStmt;
} IfStat;

/**
 * A match clause for a match statement
 */
typedef struct {
	Expression *expr;
	Block *body;
} MatchClause;

/**
 * match statement
 */
typedef struct {
	Expression *expr;
	Vector *clauses;
} MatchStat;

/**
 * node representing a for loop
 */
typedef struct {
	Type *type;
	char *index;
	Vector *expr;
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
	int type;
} StructuredStatement;

/**
 * Statements, structured and unstructured
 */
typedef struct {
	StructuredStatement *structured;
	UnstructuredStatement *unstructured;
	int type;
} Statement;

Node *createNode(void *data, NodeType type);

IdentifierList *createIdentifierList();

Literal *createLiteral(char *value, int type);

TypeLit *createTypeLit();

BaseType *createBaseType();

UnaryExpr *createUnaryExpr();

ArraySubExpr *createArraySubExpr(Expression *lhand);

MemberAccessExpr *createMemberAccessExpr(Expression *rhand, char *value);

Call *createCall(Expression *rhand);

Expression *createExpression();

BinaryExpr *createBinaryExpr();

TypeName *createTypeName(char *name);

ArrayType *createArrayType(Expression *length, Type *type);

PointerType *createPointerType(BaseType *type);

FieldDecl *createFieldDecl(Type *type, bool mutable);

FieldDeclList *createFieldDeclList();

StructDecl *createStructDecl(char *name);

StatementList *createStatementList();

Block *createBlock();

ParameterSection *createParameterSection(Type *type, bool mutable);

Parameters *createParameters();

Receiver *createReceiver(Type *type, char *name, bool mutable);

FunctionSignature *createFunctionSignature(char *name, Parameters *params, bool mutable, Type *type);

FunctionDecl *createFunctionDecl();

VariableDecl *createVariableDecl(Type *type, char *name, bool mutable, Expression *rhand);

Declaration *createDeclaration();

IncDecStat *createIncDecStat(Expression *rhand, int inc);

ReturnStat *createReturnStat(Expression *rhand);

BreakStat *createBreakStat();

ContinueStat *createContinueStat();

LeaveStat *createLeaveStat();

Assignment *createAssignment(Expression *lhand, Expression *rhand);

UnstructuredStatement *createUnstructuredStatement();

ElseStat *createElseStat();

IfStat *createIfStat();

MatchClause *createMatchClause();

MatchStat *createMatchStat(Expression *rhand);

ForStat *createForStat(Type *type, char *index);

StructuredStatement *createStructuredStatement();

Statement *createStatement();

Type *createType();

void cleanupAST(Vector *nodes);

void destroyNode(Node *node);

void destroyIdentifierList(IdentifierList *list);

void destroyBaseType(BaseType *type);

void destroyLiteral(Literal *lit);

void destroyUnaryExpr(UnaryExpr *rhand);

void destroyArraySubExpr(ArraySubExpr *rhand);

void destroyMemberAccessExpr(MemberAccessExpr *rhand);

void destroyCall(Call *call);

void destroyExpression(Expression *rhand);

void destroyBinaryExpression(BinaryExpr *binary);

void destroyTypeName(TypeName *typeName);

void destroyArrayType(ArrayType *arrayType);

void destroyPointerType(PointerType *pointerType);

void destroyFieldDecl(FieldDecl *decl);

void destroyFieldDeclList(FieldDeclList *list);

void destroyStructDecl(StructDecl *decl);

void destroyStatementList(StatementList *list);

void destroyBlock(Block *block);

void destroyParameterSection(ParameterSection *param);

void destroyParameters(Parameters *params);

void destroyReceiver(Receiver *receiver);

void destroyFunctionSignature(FunctionSignature *func);

void destroyFunctionDecl(FunctionDecl *decl);

void destroyVariableDecl(VariableDecl *decl);

void destroyDeclaration(Declaration *decl);

void destroyIncDecStat(IncDecStat *stmt);

void destroyReturnStat(ReturnStat *stmt);

void destroyBreakStat(BreakStat *stmt);

void destroyContinueStat(ContinueStat *stmt);

void destroyLeaveStat(LeaveStat *stmt);

void destroyAssignment(Assignment *assign);

void destroyUnstructuredStatement(UnstructuredStatement *stmt);

void destroyElseStat(ElseStat *stmt);

void destroyIfStat(IfStat *stmt);

void destroyMatchClause(MatchClause *mclause);

void destroyMatchStat(MatchStat *match);

void destroyForStat(ForStat *stmt);

void destroyStructuredStatement(StructuredStatement *stmt);

void destroyStatement(Statement *stmt);

void destroyTypeLiteral(TypeLit *type);

void destroyType(Type *type);

#endif // AST_H
