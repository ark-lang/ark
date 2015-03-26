#ifndef AST_H
#define AST_H

#include "util.h"
#include "vector.h"

typedef struct Type s_Type;
typedef struct Expression s_Expression;
typedef struct {} ContinueStat;
typedef struct {} BreakStat;

typedef struct {
	Vector *values;
} IdentifierList;

typedef enum {
	LITERAL_NUMBER,
	LITERAL_STRING,
	LITERAL_CHARACTER
} LiteralType;

typedef struct {
	char *value;
	LiteralType type;
} Literal;

typedef struct {
	s_Expression *lhand;
	char operand;
	s_Expression *rhand;
} BinaryExpr;

typedef struct {
	char operand;
	s_Expression *rhand;
} UnaryExpr;

typedef struct {
	s_Expression *lhand;

	// optional
	s_Expression *start;
	s_Expression *end;
} ArraySubExpr;

typedef struct {
	s_Expression *expr;
	char *value;
} MemberAccessExpr;

typedef struct {
	Literal *literal;
	s_Expression *parenExpr;
	ArraySubExpr *arraySlice;
	MemberAccessExpr *memberAccess;
} PrimaryExpr;

typedef enum {
	PRIMARY_EXPR,
	BINARY_EXPR,
	UNARY_EXPR,
	MEMBER_ACCESS_EXPR,
	ARRAY_SLICE_EXPR,
} ExpressionType;

typedef struct s_Expression {
	PrimaryExpr *primary;
	BinaryExpr *binary;
	UnaryExpr *unary;
	MemberAccessExpr *memberAccess;
	ArraySubExpr *arraySlice;
} Expression;

typedef struct {
	char *name;
} TypeName;

typedef struct {
	Expression *length;
	struct s_Type *type;
} ArrayType;

typedef struct {
	struct s_Type *type;
} PointerType;

typedef struct {
	IdentifierList *idenList;
	struct s_Type *type;
	bool mutable;
} FieldDecl;

typedef struct {
	Vector *members;
} FieldDeclList;

typedef struct {
	FieldDeclList *fields;
} StructDecl;

typedef struct {
	Vector *stmts;
} StatementList;

typedef enum {
	MULTI_STATEMENT_BLOCK,
	SINGLE_STATEMENT_BLOCK
} BlockType;

typedef struct {
	StatementList *stmtList;
	BlockType type;
} Block;

typedef struct {
	s_Type *type;
	IdentifierList *identList;
	bool mutable;
} ParameterSection;

typedef struct {
	Vector *list;
} ParameterList;

typedef struct {
	Vector *paramList;
} Parameters;

typedef struct {
	s_Type *type;
	char *name;
	bool mutable;
} Receiver;

typedef struct {
	Receiver *receiver;
	char *name;
	Parameters *parameters;
} FunctionSignature;

typedef struct {
	FunctionSignature *signature;
	Block *body;
} FunctionDecl;

typedef struct {
	bool mutable;
	s_Type *type;
	char *name;
	s_Expression *expr;
} VariableDecl;

typedef struct {
	FunctionDecl *funcDecl;
	StructDecl *structDecl;
	VariableDecl *varDecl;
} Declaration;

typedef enum {
	// IOD = Increment or Decrement
	IOD_INCREMENT,
	IOD_DECREMENT,
} IncOrDec;

typedef struct {
	Expression *expr;
	IncOrDec type;
} IncDecStat;

typedef struct {
	s_Expression *expr;
} ReturnStat;

typedef struct {
	PrimaryExpr *primary;
	s_Expression *expr;
} Assignment;

typedef struct {
	ReturnStat *retStmt;
	BreakStat *breakStmt;
	ContinueStat *conStmt;
} LeaveStat;

typedef struct {
	Declaration *decl;
	LeaveStat *leave;
	IncDecStat *incDec;
	Assignment *assignment;
} UnstructuredStatement;

typedef struct {
	Block *body;
} ElseStat;

typedef struct {
	s_Expression *expr;
	Block *body;
	ElseStat *elseStmt;
} IfStat;

typedef struct {
	s_Expression *expr;
	Block *body;
	LeaveStat *leave;
} MatchClause;

typedef struct {
	s_Expression *expr;
	Vector *clauses;
} MatchStat;

typedef struct {
	s_Type *type;
	char *index;
	PrimaryExpr *start, *end, *step;
	Block *body;
} ForStat;

typedef struct {
	Block *block;
	IfStat *ifStmt;
	MatchStat *matchStmt;
	ForStat *forStmt;
} StructuredStatement;

typedef struct {
	StructuredStatement *structured;
	UnstructuredStatement *unstructured;
} Statement;

typedef struct s_Type {
	TypeName *typeName;
	ArrayType *arrayType;
	PointerType *pointerType;
	StructDecl *structType;
} Type;

#endif // AST_H
