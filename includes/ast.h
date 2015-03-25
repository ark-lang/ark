#ifndef AST_H
#define AST_H

#include "util.h"
#include "vector.h"

typedef struct Type s_Type;
typedef struct Expression s_Expression;

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

typedef enum {
	BINARY_EXPR,
	UNARY_EXPR,
	PRIMARY_EXPR,
} ExpressionType;

typedef struct {
	s_Expression *lhand;
	char operand;
	s_Expression *rhand;
} BinaryExpression;

typedef struct {
	char operand;
	s_Expression *rhand;
} UnaryExpression;

typedef struct {
	s_Expression *lhand;

	// optional
	s_Expression *start;
	s_Expression *end;
} ArraySliceExpression;

typedef struct {

} MemberAccessExpression;

typedef struct {
	Literal *literal;
	s_Expression *parenExpr;
	ArraySliceExpression *arraySlice;
	MemberAccessExpression *memberAccess;
} PrimaryExpr;

typedef struct s_Expression {

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
} FieldDecl;

typedef struct {
	Vector *members;
} FieldDeclList;

typedef struct {
	FieldDeclList *fields;
} StructType;

typedef struct {

} FunctionType;

typedef struct s_Type {
	TypeName *typeName;
	ArrayType *arrayType;
	PointerType *pointerType;
	StructType *structType;
} Type;

IdentifierList *createIdentifierList();

void destroyIdentifierList(IdentifierList *list);

Expression *createExpression();

void destroyExpression(Expression *expr);

TypeName *createTypeName();

void destroyTypeName(TypeName *typeName);

ArrayType *createArrayType();

void destroyArrayType(ArrayType *arrayType);

PointerType *createPointerType();

void destroyPointerType(PointerType *pointerType);

FieldDecl *createFieldDecl();

void destroyFieldDecl(FieldDecl *fieldDecl);

FieldDeclList *createFieldDeclList();

void destroyFieldDeclList(FieldDeclList *fieldDeclList);

StructType *createStructType();

void destroyStructType(StructType *structType);

FunctionType *createFunctionType();

void destroyFunctionType(FunctionType *funcType);

Type *createType();

void destroyType(Type *type);

#endif // AST_H
