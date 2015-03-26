#include "ast.h"

IdentifierList *createIdentifierList() {
	return safeMalloc(sizeof(IdentifierList));
}

Literal *createLiteral(char *value, LiteralType type) {
	Literal *lit = safeMalloc(sizeof(*lit));
	lit->value = value;
	lit->type = type;
	return lit;
}

BinaryExpr *createBinaryExpr(s_Expression *lhand, char op, s_Expression *rhand) {
	BinaryExpr *expr = safeMalloc(sizeof(*expr));
	expr->lhand = lhand;
	expr->operand = op;
	expr->rhand = rhand;
	return expr;
}

UnaryExpr *createUnaryExpr(char operand, s_Expression *rhand) {
	UnaryExpr *unary = safeMalloc(sizeof(*unary));
	unary->operand = operand;
	unary->rhand = rhand;
	return unary;
}

ArraySubExpr *createArraySubExpr(s_Expression *lhand) {
	ArraySubExpr *expr = safeMalloc(sizeof(*expr));
	expr->lhand = lhand;
	return expr;
}

MemberAccessExpr *createMemberAccessExpr(s_Expression *expr, char *value) {
	MemberAccessExpr *mem = safeMalloc(sizeof(*mem));
	mem->expr = expr;
	mem->value = value;
	return mem;
}

PrimaryExpr *createPrimaryExpr() {
	return safeMalloc(sizeof(PrimaryExpr));
}

Expression *createExpression() {
	return safeMalloc(sizeof(Expression));
}

TypeName *createTypeName(char *name) {
	TypeName *type = safeMalloc(sizeof(*type));
	type->name = name;
	return type;
}

ArrayType *createArrayType(Expression *length, struct s_Type *type) {
	ArrayType *arr = safeMalloc(sizeof(*arr));
	arr->length = length;
	arr->type = type;
	return arr;
}

PointerType *createPointerType(struct s_Type *type) {
	PointerType *pointer = safeMalloc(sizeof(*pointer));
	pointer->type = type;
	return pointer;
}

FieldDecl *createFieldDecl(struct s_Type *type, bool mutable) {
	FieldDecl *field = safeMalloc(sizeof(*field));
	field->type = type;
	field->mutable = mutable;
	return field;
}

FieldDeclList *createFieldDeclList() {
	return safeMalloc(sizeof(FieldDeclList));
}

StructDecl *createStructDecl(char *name) {
	StructDecl *str = safeMalloc(sizeof(*str));
	str->name = name;
	return str;
}

StatementList *createStatementList() {
	return safeMalloc(sizeof(StatementList));
}

Block *createBlock() {
	return safeMalloc(sizeof(Block));
}

ParameterSection *createParameterSection(s_Type *type, bool mutable) {
	ParameterSection *param = safeMalloc(sizeof(*param));
	param->type = type;
	param->mutable = mutable;
	return param;
}

Parameters *createParameters() {
	return safeMalloc(sizeof(Parameters));
}

Receiver *createReceiver(s_Type *type, char *name, bool mutable) {
	Receiver *receiver = safeMalloc(sizeof(*receiver));
	receiver->type = type;
	receiver->name = name;
	receiver->mutable = mutable;
	return receiver;
}

FunctionSignature *createFunctionSignature(char *name, Parameters *params, bool mutable) {
	FunctionSignature *func = safeMalloc(sizeof(*func));
	func->name = name;
	func->parameters = params;
	func->mutable = mutable;
	return func;
}

FunctionDecl *createFunctionDecl() {
	return safeMalloc(sizeof(FunctionDecl));
}

VariableDecl *createVariableDecl(s_Type *type, char *name, bool mutable, s_Expression *expr) {
	VariableDecl *var = safeMalloc(sizeof(*var));
	var->type = type;
	var->name = name;
	var->mutable = mutable;
	var->expr = expr;
	return var;
}

Declaration *createDeclaration() {
	return safeMalloc(sizeof(Declaration));
}

IncDecStat *createIncDecStat(s_Expression *expr, IncOrDec type) {
	IncDecStat *inc = safeMalloc(sizeof(*inc));
	inc->expr = expr;
	inc->type = type;
	return inc;
}

ReturnStat *createReturnStat(s_Expression *expr) {
	ReturnStat *ret = safeMalloc(sizeof(*ret));
	ret->expr = expr;
	return ret;
}

BreakStat *createBreakStat() {
	return safeMalloc(sizeof(BreakStat));
}

ContinueStat *createContinueStat() {
	return safeMalloc(sizeof(ContinueStat));
}

LeaveStat *createLeaveStat() {
	return safeMalloc(sizeof(LeaveStat));
}

Assignment *createAssignment(PrimaryExpr *primaryExpr, s_Expression *expr) {
	Assignment *assign = safeMalloc(sizeof(*assign));
	assign->primary = primaryExpr;
	assign->expr = expr;
	return assign;
}

UnstructuredStatement *createUnstructuredStatement() {
	return safeMalloc(sizeof(UnstructuredStatement));
}

ElseStat *createElseStat() {
	return safeMalloc(sizeof(ElseStat));
}

IfStat *createIfStat() {
	return safeMalloc(sizeof(IfStat));
}

MatchClause *createMatchClause() {
	return safeMalloc(sizeof(MatchClause));
}

MatchStat *createMatchStat(s_Expression *expr) {
	MatchStat *match = safeMalloc(sizeof(*match));
	match->expr = expr;
	return match;
}

ForStat *createForStat(s_Type *type, char *index, PrimaryExpr *start, PrimaryExpr *end) {
	ForStat *forStat = safeMalloc(sizeof(*forStat));
	forStat->type = type;
	forStat->index = index;
	forStat->start = start;
	forStat->end = end;
	return forStat;
}

StructuredStatement *createStructuredStatement() {
	return safeMalloc(sizeof(StructuredStatement));
}

Statement *createStatement() {
	return safeMalloc(sizeof(Statement));
}

Type *createType() {
	return safeMalloc(sizeof(Type));
}

