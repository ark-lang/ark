#include "ast.h"

IdentifierList *createIdentifierList() {
	IdentifierList *iden = safeMalloc(sizeof(*iden));
	iden->values = createVector(VECTOR_EXPONENTIAL);
	return iden;
}

Literal *createLiteral(char *value, int type) {
	Literal *lit = safeMalloc(sizeof(*lit));
	lit->value = value;
	lit->type = type;
	return lit;
}

TypeLit *createTypeLit() {
	return safeMalloc(sizeof(TypeLit));
}

BaseType *createBaseType() {
	return safeMalloc(sizeof(BaseType));
}

UnaryExpr *createUnaryExpr() {
	UnaryExpr *unary = safeMalloc(sizeof(*unary));
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

Call *createCall(Expression *expr) {
	Call *call = safeMalloc(sizeof(*call));
	call->arguments = createVector(VECTOR_EXPONENTIAL);
	call->callee = expr;
	return call;
}

Expression *createExpression() {
	return safeMalloc(sizeof(Expression));
}

BinaryExpr *createBinaryExpr() {
	return safeMalloc(sizeof(BinaryExpr));
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

PointerType *createPointerType(BaseType *type) {
	PointerType *pointer = safeMalloc(sizeof(*pointer));
	pointer->baseType = type;
	return pointer;
}

FieldDecl *createFieldDecl(Type *type, bool mutable) {
	FieldDecl *field = safeMalloc(sizeof(*field));
	field->type = type;
	field->mutable = mutable;
	return field;
}

FieldDeclList *createFieldDeclList() {
	FieldDeclList *fieldDeclList = safeMalloc(sizeof(*fieldDeclList));
	fieldDeclList->members = createVector(VECTOR_EXPONENTIAL);
	return fieldDeclList;
}

StructDecl *createStructDecl(char *name) {
	StructDecl *str = safeMalloc(sizeof(*str));
	str->name = name;
	return str;
}

StatementList *createStatementList() {
	StatementList *stmtList = safeMalloc(sizeof(*stmtList));
	stmtList->stmts = createVector(VECTOR_EXPONENTIAL);
	return stmtList;
}

Block *createBlock() {
	Block *block = safeMalloc(sizeof(*block));
	block->stmtList = createStatementList();
	return block;
}

ParameterSection *createParameterSection(Type *type, bool mutable) {
	ParameterSection *param = safeMalloc(sizeof(*param));
	param->type = type;
	param->mutable = mutable;
	return param;
}

Parameters *createParameters() {
	Parameters *params = safeMalloc(sizeof(*params));
	params->paramList = createVector(VECTOR_EXPONENTIAL);
	return params;
}

Receiver *createReceiver(Type *type, char *name, bool mutable) {
	Receiver *receiver = safeMalloc(sizeof(*receiver));
	receiver->type = type;
	receiver->name = name;
	receiver->mutable = mutable;
	return receiver;
}

FunctionSignature *createFunctionSignature(char *name, Parameters *params,
		bool mutable, Type *type) {
	FunctionSignature *func = safeMalloc(sizeof(*func));
	func->name = name;
	func->parameters = params;
	func->mutable = mutable;
	func->type = type;
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

IncDecStat *createIncDecStat(Expression *expr, int amount) {
	IncDecStat *inc = safeMalloc(sizeof(*inc));
	inc->expr = expr;
	inc->amount = amount;
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

Assignment *createAssignment(Expression *lhand, Expression *rhand) {
	Assignment *assign = safeMalloc(sizeof(*assign));
	assign->primary = lhand;
	assign->expr = rhand;
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
	match->clauses = createVector(VECTOR_EXPONENTIAL);
	return match;
}

ForStat *createForStat(Type *type, char *index) {
	ForStat *forStat = safeMalloc(sizeof(*forStat));
	forStat->type = type;
	forStat->index = index;
	forStat->expr = createVector(VECTOR_EXPONENTIAL);
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
		case UNARY_EXPR_NODE: destroyUnaryExpr(node->data); break;
		case ARRAY_SUB_EXPR_NODE: destroyArraySubExpr(node->data); break;
		case MEMBER_ACCESS_NODE: destroyMemberAccessExpr(node->data); break;
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
	if (!list) return;
	destroyVector(list->values);
	free(list);
}

void destroyBaseType(BaseType *type) {
	if (!type) return;
	destroyTypeName(type->type);
	free(type);
}

void destroyLiteral(Literal *lit) {
	if (!lit) return;
	free(lit);
}

void destroyUnaryExpr(UnaryExpr *expr) {
	if (!expr) return;
	destroyExpression(expr->lhand);
	free(expr);
}

void destroyArraySubExpr(ArraySubExpr *expr) {
	if (!expr) return;
	destroyExpression(expr->lhand);
	destroyExpression(expr->start);
	destroyExpression(expr->end);
	free(expr);
}

void destroyMemberAccessExpr(MemberAccessExpr *expr) {
	if (!expr) return;
	destroyExpression(expr->expr);
	free(expr);
}

void destroyCall(Call *call) {
	if (!call) return;
	for (int i = 0; i < call->arguments->size; i++) {
		destroyExpression(getVectorItem(call->arguments, i));
	}
	destroyVector(call->arguments);
	destroyExpression(call->callee);
	free(call);
}

void destroyExpression(Expression *expr) {
	if (!expr) return;
	destroyBinaryExpression(expr->binary);
	destroyCall(expr->call);
	destroyLiteral(expr->lit);
	destroyType(expr->type);
	destroyUnaryExpr(expr->unary);
	free(expr);
}

void destroyBinaryExpression(BinaryExpr *binary) {
	if (!binary) return;
	destroyExpression(binary->lhand);
	destroyExpression(binary->rhand);
	free(binary);
}

void destroyTypeName(TypeName *typeName) {
	if (!typeName) return;
	free(typeName);
}

void destroyArrayType(ArrayType *arrayType) {
	if (!arrayType) return;
	destroyExpression(arrayType->length);
	destroyType(arrayType->type);
	free(arrayType);
}

void destroyPointerType(PointerType *pointerType) {
	if (!pointerType) return;
	destroyBaseType(pointerType->baseType);
	free(pointerType);
}

void destroyFieldDecl(FieldDecl *decl) {
	if (!decl) return;
	destroyType(decl->type);
	free(decl);
}

void destroyFieldDeclList(FieldDeclList *list) {
	if (!list) return;
	destroyVector(list->members);
	free(list);
}

void destroyStructDecl(StructDecl *decl) {
	if (!decl) return;
	destroyFieldDeclList(decl->fields);
	free(decl);
}

void destroyStatementList(StatementList *list) {
	if (!list) return;
	for (int i = 0; i < list->stmts->size; i++) {
		destroyStatement(getVectorItem(list->stmts, i));
	}
	destroyVector(list->stmts);
	free(list);
}

void destroyBlock(Block *block) {
	if (!block) return;
	destroyStatementList(block->stmtList);
	free(block);
}

void destroyParameterSection(ParameterSection *param) {
	if (!param) return;
	destroyType(param->type);
	free(param);
}

void destroyParameters(Parameters *params) {
	if (!params) return;
	for (int i = 0; i < params->paramList->size; i++) {
		destroyParameterSection(getVectorItem(params->paramList, i));
	}
	destroyVector(params->paramList);
	free(params);
}

void destroyReceiver(Receiver *receiver) {
	if (!receiver) return;
	destroyType(receiver->type);
	free(receiver);
}

void destroyFunctionSignature(FunctionSignature *func) {
	if (!func) return;
	destroyReceiver(func->receiver);
	destroyParameters(func->parameters);
	free(func);
}

void destroyFunctionDecl(FunctionDecl *decl) {
	if (!decl) return;
	destroyBlock(decl->body);
	destroyFunctionSignature(decl->signature);
	free(decl);
}

void destroyVariableDecl(VariableDecl *decl) {
	if (!decl) return;
	destroyExpression(decl->expr);
	destroyType(decl->type);
	free(decl);
}

void destroyDeclaration(Declaration *decl) {
	if (!decl) return;
	destroyFunctionDecl(decl->funcDecl);
	destroyStructDecl(decl->structDecl);
	destroyVariableDecl(decl->varDecl);
	free(decl);
}

void destroyIncDecStat(IncDecStat *stmt) {
	if (!stmt) return;
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyReturnStat(ReturnStat *stmt) {
	if (!stmt) return;
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyBreakStat(BreakStat *stmt) {
	if (!stmt) return;
	free(stmt);
}

void destroyContinueStat(ContinueStat *stmt) {
	if (!stmt) return;
	free(stmt);
}

void destroyLeaveStat(LeaveStat *stmt) {
	if (!stmt) return;
	destroyBreakStat(stmt->breakStmt);
	destroyReturnStat(stmt->retStmt);
	destroyContinueStat(stmt->conStmt);
	free(stmt);
}

void destroyAssignment(Assignment *assign) {
	if (!assign) return;
	destroyExpression(assign->expr);
	destroyExpression(assign->primary);
	free(assign);
}

void destroyUnstructuredStatement(UnstructuredStatement *stmt) {
	if (!stmt) return;
	switch (stmt->type) {
		case ASSIGNMENT_NODE: destroyAssignment(stmt->assignment); break;
		case DECLARATION_NODE: destroyDeclaration(stmt->decl); break;
		case INC_DEC_STAT_NODE: destroyIncDecStat(stmt->incDec); break;
		case LEAVE_STAT_NODE: destroyLeaveStat(stmt->leave); break;
	}
	free(stmt);
}

void destroyElseStat(ElseStat *stmt) {
	if (!stmt) return;
	destroyBlock(stmt->body);
	free(stmt);
}

void destroyIfStat(IfStat *stmt) {
	if (!stmt) return;
	destroyBlock(stmt->body);
	destroyElseStat(stmt->elseStmt);
	destroyExpression(stmt->expr);
	free(stmt);
}

void destroyMatchClause(MatchClause *mclause) {
	if (!mclause) return;
	destroyBlock(mclause->body);
	destroyExpression(mclause->expr);
	free(mclause);
}

void destroyMatchStat(MatchStat *match) {
	if (!match) return;
	for (int i = 0; i < match->clauses->size; i++) {
		destroyMatchClause(getVectorItem(match->clauses, i));
	}
	destroyVector(match->clauses);
	destroyExpression(match->expr);
	free(match);
}

void destroyForStat(ForStat *stmt) {
	if (!stmt) return;
	destroyBlock(stmt->body);
	for (int i = 0 ; i < stmt->expr->size; i++) {
		destroyExpression(getVectorItem(stmt->expr, i));
	}
	destroyVector(stmt->expr);
	destroyType(stmt->type);
	free(stmt);
}

void destroyStructuredStatement(StructuredStatement *stmt) {
	if (!stmt) return;
	destroyBlock(stmt->block);
	destroyForStat(stmt->forStmt);
	destroyIfStat(stmt->ifStmt);
	destroyMatchStat(stmt->matchStmt);
	free(stmt);
}

void destroyStatement(Statement *stmt) {
	if (!stmt) return;
	destroyStructuredStatement(stmt->structured);
	destroyUnstructuredStatement(stmt->unstructured);
	free(stmt);
}

void destroyTypeLiteral(TypeLit *type) {
	if (!type) return;
	destroyArrayType(type->arrayType);
	destroyPointerType(type->pointerType);
	free(type);
}

void destroyType(Type *type) {
	if (!type) return;
	destroyTypeLiteral(type->typeLit);
	destroyTypeName(type->typeName);
	free(type);
}
