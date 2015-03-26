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

BinaryExpr *createBinaryExpr(Expression *lhand, char op, Expression *rhand) {
	BinaryExpr *expr = safeMalloc(sizeof(*expr));
	expr->lhand = lhand;
	expr->operand = op;
	expr->rhand = rhand;
	return expr;
}

UnaryExpr *createUnaryExpr(char operand, Expression *rhand) {
	UnaryExpr *unary = safeMalloc(sizeof(*unary));
	unary->operand = operand;
	unary->rhand = rhand;
	return unary;
}

ArraySubExpr *createArraySubExpr(Expression *lhand) {
	ArraySubExpr *expr = safeMalloc(sizeof(*expr));
	expr->lhand = lhand;
	return expr;
}

MemberAccessExpr *createMemberAccessExpr(Expression *expr, char *value) {
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

ArrayType *createArrayType(Expression *length, Type *type) {
	ArrayType *arr = safeMalloc(sizeof(*arr));
	arr->length = length;
	arr->type = type;
	return arr;
}

PointerType *createPointerType(Type *type) {
	PointerType *pointer = safeMalloc(sizeof(*pointer));
	pointer->type = type;
	return pointer;
}

FieldDecl *createFieldDecl(Type *type, bool mutable) {
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

ParameterSection *createParameterSection(Type *type, bool mutable) {
	ParameterSection *param = safeMalloc(sizeof(*param));
	param->type = type;
	param->mutable = mutable;
	return param;
}

Parameters *createParameters() {
	return safeMalloc(sizeof(Parameters));
}

Receiver *createReceiver(Type *type, char *name, bool mutable) {
	Receiver *receiver = safeMalloc(sizeof(*receiver));
	receiver->type = type;
	receiver->name = name;
	receiver->mutable = mutable;
	return receiver;
}

FunctionSignature *createFunctionSignature(char *name, Parameters *params,
		bool mutable) {
	FunctionSignature *func = safeMalloc(sizeof(*func));
	func->name = name;
	func->parameters = params;
	func->mutable = mutable;
	return func;
}

FunctionDecl *createFunctionDecl() {
	return safeMalloc(sizeof(FunctionDecl));
}

VariableDecl *createVariableDecl(Type *type, char *name, bool mutable,
		Expression *expr) {
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

IncDecStat *createIncDecStat(Expression *expr, IncOrDec type) {
	IncDecStat *inc = safeMalloc(sizeof(*inc));
	inc->expr = expr;
	inc->type = type;
	return inc;
}

ReturnStat *createReturnStat(Expression *expr) {
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

Assignment *createAssignment(PrimaryExpr *primaryExpr, Expression *expr) {
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

MatchStat *createMatchStat(Expression *expr) {
	MatchStat *match = safeMalloc(sizeof(*match));
	match->expr = expr;
	return match;
}

ForStat *createForStat(Type *type, char *index, PrimaryExpr *start,
		PrimaryExpr *end) {
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

void cleanupAST(Vector *nodes) {
	int i;
	for (i = 0; i < nodes->size; i++) {
		Node *node = getVectorItem(nodes, i);

		switch (node->type) {
		case IDENTIFIER_LIST_NODE: destroyIdentifierList(node->data); break;
		case LITERAL_NODE: destroyLiteral(node->data); break;
		case BINARY_EXPR_NODE: destroyBinaryExpr(node->data); break;
		case UNARY_EXPR_NODE: destroyUnaryExpr(node->data); break;
		case ARRAY_SUB_EXPR_NODE: destroyArraySubExpr(node->data); break;
		case MEMBER_ACCESS_NODE: destroyMemberAccessExpr(node->data); break;
		case PRIMARY_EXPR_NODE: destroyPrimaryExpr(node->data); break;
		case EXPR_NODE: destroyExpression(node->data); break;
		case TYPE_NAME_NODE: destroyTypeName(node->data); break;
		case ARRAY_TYPE_NODE: destroyArrayType(node->data); break;
		case POINTER_TYPE_NODE: destroyPointerType(node->data); break;
		case FIELD_DECL_NODE: destroyFieldDecl(node->data); break;
		case FIELD_DECL_LIST_NODE: destroyFieldDeclList(node->data); break;
		case STRUCT_DECL_NODE: destroyStructDecl(node->data); break;
		case STATEMENT_LIST_NODE: destroyStatementList(node->data); break;
		case BLOCK_NODE: destroyBlock(node->data); break;
		case PARAMETER_SECTION_NODE: destroyParameterSection(node->data); break;
		case PARAMETERS_NODE: destroyParameters(node->data); break;
		case RECEIVER_NODE: destroyReceiver(node->data); break;
		case FUNCTION_SIGNATURE_NODE: destroyFunctionSignature(node->data); break;
		case FUNCTION_DECL_NODE: destroyFunctionDecl(node->data); break;
		case VARIABLE_DECL_NODE: destroyVariableDecl(node->data); break;
		case DECLARATION_NODE: destroyDeclaration(node->data); break;
		case INC_DEC_STAT_NODE: destroyIncDecStat(node->data); break;
		case RETURN_STAT_NODE: destroyReturnStat(node->data); break;
		case BREAK_STAT_NODE: destroyBreakStat(node->data); break;
		case CONTINUE_STAT_NODE: destroyContinueStat(node->data); break;
		case LEAVE_STAT_NODE: destroyLeaveStat(node->data); break;
		case ASSIGNMENT_NODE: destroyAssignment(node->data); break;
		case UNSTRUCTURED_STATEMENT_NODE: destroyUnstructuredStatement(node->data); break;
		case ELSE_STAT_NODE: destroyElseStat(node->data); break;
		case IF_STAT_NODE: destroyIfStat(node->data); break;
		case MATCH_CLAUSE_STAT: destroyMatchClause(node->data); break;
		case MATCH_STAT_NODE: destroyMatchStat(node->data); break;
		case FOR_STAT_NODE: destroyForStat(node->data); break;
		case STRUCTURED_STATEMENT_NODE: destroyStructuredStatement(node->data); break;
		case STATEMENT_NODE: destroyStatement(node->data); break;
		case TYPE_NODE: destroyType(node->data); break;
		default: printf("un-recognized node %d\n", node->type); break;
		}
	}
}

void destroyIdentifierList(IdentifierList *list) {
	destroyVector(list->values);
	free(list);
}

void destroyLiteral(Literal *lit) {
	free(lit);
}

void destroyBinaryExpr(BinaryExpr *expr) {
	destroyExpression(expr->lhand);
	destroyExpression(expr->rhand);
	free(expr);
}

void destroyUnaryExpr(UnaryExpr *expr) {
	destroyExpression(expr->rhand);
	free(expr);
}

void destroyArraySubExpr(ArraySubExpr *expr) {
	destroyExpression(expr->lhand);
	destroyExpression(expr->start);
	destroyExpression(expr->end);
	free(expr);
}

void destroyMemberAccessExpr(MemberAccessExpr *expr) {
	destroyExpression(expr->expr);
	free(expr);
}

void destroyPrimaryExpr(PrimaryExpr *expr) {
	destroyArraySubExpr(expr->arraySlice);
	destroyLiteral(expr->literal);
	destroyMemberAccessExpr(expr->memberAccess);
	destroyExpression(expr->parenExpr);
	free(expr);
}

void destroyExpression(Expression *expr) {
	destroyPrimaryExpr(expr->primary);
	destroyBinaryExpr(expr->binary);
	destroyUnaryExpr(expr->unary);
	destroyMemberAccessExpr(expr->memberAccess);
	destroyArraySubExpr(expr->arraySlice);
	free(expr);
}

void destroyTypeName(TypeName *typeName) {
	free(typeName);
}

void destroyArrayType(ArrayType *arrayType) {
	destroyExpression(arrayType->length);
	destroyType(arrayType->type);
	free(arrayType);
}

void destroyPointerType(PointerType *pointerType) {
	destroyType(pointerType->type);
	free(pointerType);
}

void destroyFieldDecl(FieldDecl *decl) {
	destroyIdentifierList(decl->idenList);
	destroyType(decl->type);
	free(decl);
}

void destroyFieldDeclList(FieldDeclList *list) {
	destroyVector(list->members);
	free(list);
}

void destroyStructDecl(StructDecl *decl) {
	destroyFieldDeclList(decl->fields);
	free(decl);
}

void destroyStatementList(StatementList *list) {
	destroyVector(list->stmts);
	free(list);
}

void destroyBlock(Block *block) {
	destroyStatementList(block->stmtList);
	free(block);
}

void destroyParameterSection(ParameterSection *param) {
	destroyType(param->type);
	free(param);
}

void destroyParameters(Parameters *params) {
	destroyVector(params->paramList);
	free(params);
}

void destroyReceiver(Receiver *receiver) {
	destroyType(receiver->type);
	free(receiver);
}

void destroyFunctionSignature(FunctionSignature *func) {
	destroyReceiver(func->receiver);
	destroyParameters(func->parameters);
	free(func);
}

void destroyFunctionDecl(FunctionDecl *decl) {
	destroyBlock(decl->body);
	destroyFunctionSignature(decl->signature);
	free(decl);
}

void destroyVariableDecl(VariableDecl *decl) {
	destroyExpression(decl->expr);
	destroyType(decl->type);
	free(decl);
}

void destroyDeclaration(Declaration *decl) {
	destroyFunctionDecl(decl->funcDecl);
	destroyStructDecl(decl->structDecl);
	destroyVariableDecl(decl->varDecl);
	free(decl);
}

void destroyIncDecStat(IncDecStat *stmt) {
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyReturnStat(ReturnStat *stmt) {
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyBreakStat(BreakStat *stmt) {
	free(stmt);
}

void destroyContinueStat(ContinueStat *stmt) {
	free(stmt);
}

void destroyLeaveStat(LeaveStat *stmt) {
	destroyBreakStat(stmt->breakStmt);
	destroyReturnStat(stmt->retStmt);
	destroyContinueStat(stmt->conStmt);
	free(stmt);
}

void destroyAssignment(Assignment *assign) {
	destroyExpression(assign->expr);
	destroyPrimaryExpr(assign->primary);
	free(assign);
}

void destroyUnstructuredStatement(UnstructuredStatement *stmt) {
	destroyAssignment(stmt->assignment);
	destroyDeclaration(stmt->decl);
	destroyIncDecStat(stmt->incDec);
	destroyLeaveStat(stmt->leave);
	free(stmt);
}

void destroyElseStat(ElseStat *stmt) {
	destroyBlock(stmt->body);
	free(stmt);
}

void destroyIfStat(IfStat *stmt) {
	destroyBlock(stmt->body);
	destroyElseStat(stmt->elseStmt);
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyMatchClause(MatchClause *mclause) {
	destroyBlock(mclause->body);
	destroyExpression(mclause->expr);
	destroyLeaveStat(mclause->leave);
	free(mclause);
}

void destroyMatchStat(MatchStat *match) {
	destroyVector(match->clauses);
	destroyExpression(match->expr);
	free(match);
}

void destroyForStat(ForStat *stmt) {
	destroyBlock(stmt->body);
	destroyPrimaryExpr(stmt->end);
	destroyPrimaryExpr(stmt->start);
	destroyPrimaryExpr(stmt->step);
	destroyType(stmt->type);
	free(stmt);
}

void destroyStructuredStatement(StructuredStatement *stmt) {
	destroyBlock(stmt->block);
	destroyForStat(stmt->forStmt);
	destroyIfStat(stmt->ifStmt);
	destroyMatchStat(stmt->matchStmt);
	free(stmt);
}

void destroyStatement(Statement *stmt) {
	destroyStructuredStatement(stmt->structured);
	destroyUnstructuredStatement(stmt->unstructured);
}

void destroyType(Type *type) {
	destroyArrayType(type->arrayType);
	destroyPointerType(type->pointerType);
	destroyTypeName(type->typeName);
}
