#include "ast.h"

// Note for create functions:
// If you do not initialise the member pointers of the
// struct you are returning, (ie. self->point = NULL),
// use safeCalloc() to allocate memory for it.
// This prevents possible unsafe memory access,
// and shuts Valgrind up about uninitialised values.

/**
 * Node names corresponding to their enumerated counterpart
 */
const char *NODE_NAME[] = {
	"IDENTIFIER_LIST_NODE", "IDENTIFIER_NODE", "LITERAL_NODE", "BINARY_EXPR_NODE",
	"PRIMARY_EXPR_NODE", "EXPR_NODE", "TYPE_NAME_NODE", "TYPE_LIT_NODE", "PAREN_EXPR_NODE",
	"ARRAY_TYPE_NODE", "POINTER_TYPE_NODE", "FIELD_DECL_NODE", "UNARY_EXPR_NODE",
	"FIELD_DECL_LIST_NODE", "STRUCT_DECL_NODE", "STATEMENT_LIST_NODE",
	"BLOCK_NODE", "PARAMETER_SECTION_NODE", "PARAMETERS_NODE", "IMPL_NODE", "ENUM_DECL_NODE",
	"FUNCTION_SIGNATURE_NODE", "FUNCTION_DECL_NODE", "VARIABLE_DECL_NODE", "FUNCTION_CALL_NODE",
	"DECLARATION_NODE", "INC_DEC_STAT_NODE", "RETURN_STAT_NODE", "BREAK_STAT_NODE",
	"CONTINUE_STAT_NODE", "LEAVE_STAT_NODE", "ASSIGNMENT_NODE", "UNSTRUCTURED_STATEMENT_NODE",
	"ELSE_STAT_NODE", "IF_STAT_NODE", "MATCH_CLAUSE_STAT", "MATCH_STAT_NODE", "FOR_STAT_NODE",
	"STRUCTURED_STATEMENT_NODE", "STATEMENT_NODE", "TYPE_NODE", "POINTER_FREE_NODE", "TUPLE_TYPE_NODE",
	"TUPLE_EXPR_NODE", "OPTION_TYPE_NODE", 
	"MACRO_NODE", "USE_MACRO_NODE", "LINKER_FLAG_MACRO_NODE", "EXPR_STAT_NODE", "ARRAY_INITIALIZER_NODE",
	"CHAR_LITERAL_NODE", "OTHER_LITERAL_NODE"
};

// Useful for debugging
const char *getNodeTypeName(NodeType type) {
	if (type > NUM_OF_NODES - 1)
		errorMessage("Tried to get the name of a non-existant NodeType: %d", type);
	return NODE_NAME[type];
}

UseMacro *createUseMacro() {
	UseMacro *use = safeMalloc(sizeof(*use));
	use->files = createVector(VECTOR_EXPONENTIAL);
	return use;
}

ArrayInitializer *createArrayInitializer() {
	ArrayInitializer *arr = safeMalloc(sizeof(*arr));
	arr->values = createVector(VECTOR_EXPONENTIAL);
	return arr;
}

LinkerFlagMacro *createLinkerFlagMacro(char *flag) {
	LinkerFlagMacro *linker = safeMalloc(sizeof(*linker));
	linker->flag = flag;
	return linker;
}

Impl *createImpl(char *name, char *as) {
	Impl *impl = safeMalloc(sizeof(*impl));
	impl->name = name;
	impl->as = as;
	impl->funcs = createVector(VECTOR_EXPONENTIAL);
	return impl;
}

IdentifierList *createIdentifierList() {
	IdentifierList *iden = safeMalloc(sizeof(*iden));
	iden->values = createVector(VECTOR_EXPONENTIAL);
	return iden;
}

Literal *createLiteral() {
	Literal *lit = safeMalloc(sizeof(*lit));
	return lit;
}

OtherLit *createOtherLit(char *value) {
	OtherLit *lit = safeMalloc(sizeof(*lit));
	lit->value = value;
	return lit;
}

CharLit *createCharLit(int value) {
	CharLit *lit = safeMalloc(sizeof(*lit));
	lit->value = value;
	return lit;
}

TypeLit *createTypeLit() {
	return safeCalloc(sizeof(TypeLit));
}

UnaryExpr *createUnaryExpr() {
	UnaryExpr *unary = safeMalloc(sizeof(*unary));
	return unary;
}

TupleType *createTupleType() {
	TupleType *tuple = safeMalloc(sizeof(*tuple));
	tuple->types = createVector(VECTOR_EXPONENTIAL);
	return tuple;
}

OptionType *createOptionType(Type *type) {
	OptionType *option = safeMalloc(sizeof(*option));
	option->type = type;
	return option;
}

TupleExpr *createTupleExpr() {
	TupleExpr *tuple = safeMalloc(sizeof(*tuple));
	tuple->values = createVector(VECTOR_EXPONENTIAL);
	return tuple;
}

EnumItem *createEnumItem(char *name) {
	EnumItem *item = safeMalloc(sizeof(*item));
	item->name = name;
	return item;
}

EnumDecl *createEnumDecl(char *name) {
	EnumDecl *decl = safeMalloc(sizeof(*decl));
	decl->name = name;
	decl->items = createVector(VECTOR_EXPONENTIAL);
	return decl;
}

Call *createCall(Vector *callee) {
	Call *call = safeMalloc(sizeof(*call));
	call->arguments = createVector(VECTOR_EXPONENTIAL);
	call->callee = callee;
	return call;
}

Expression *createExpression() {
	return safeCalloc(sizeof(Expression));
}

BinaryExpr *createBinaryExpr() {
	return safeCalloc(sizeof(BinaryExpr));
}

TypeName *createTypeName(char *name) {
	TypeName *type = safeMalloc(sizeof(*type));
	type->name = name;
	return type;
}

ArrayType *createArrayType(Type *type) {
	ArrayType *arr = safeMalloc(sizeof(*arr));
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
	block->singleStatementBlock = false;
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
	params->variadic = false;
	return params;
}

FunctionSignature *createFunctionSignature(char *name, Parameters *params,
		bool mutable, Type *type) {
	FunctionSignature *func = safeMalloc(sizeof(*func));
	func->name = name;
	func->parameters = params;
	func->mutable = mutable;
	func->type = type;
	func->owner = NULL;
	func->ownerArg = NULL;
	return func;
}

FunctionDecl *createFunctionDecl() {
	FunctionDecl *func = safeMalloc(sizeof(FunctionDecl));
	func->signature = NULL;
	func->body = NULL;
	func->numOfRequiredArgs = 0;
	func->prototype = false;
	return func;
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
	return safeCalloc(sizeof(Declaration));
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
	return safeCalloc(sizeof(BreakStat));
}

ContinueStat *createContinueStat() {
	return safeCalloc(sizeof(ContinueStat));
}

LeaveStat *createLeaveStat() {
	return safeCalloc(sizeof(LeaveStat));
}

Assignment *createAssignment(char *iden, Expression *rhand) {
	Assignment *assign = safeMalloc(sizeof(*assign));
	assign->iden = iden;
	assign->expr = rhand;
	return assign;
}

UnstructuredStatement *createUnstructuredStatement() {
	return safeCalloc(sizeof(UnstructuredStatement));
}

Macro *createMacro() {
	return safeCalloc(sizeof(Macro));
}

PointerFree *createPointerFree(char *name) {
	PointerFree *pntr = safeMalloc(sizeof(*pntr));
	pntr->name = name;
	return pntr;
}

ElseStat *createElseStat() {
	return safeCalloc(sizeof(ElseStat));
}

ElseIfStat *createElseIfStat() {
	return safeCalloc(sizeof(ElseIfStat));
}

IfStat *createIfStat() {
	IfStat *ifStmt = safeMalloc(sizeof(*ifStmt));
	ifStmt->elseIfStmts = createVector(VECTOR_EXPONENTIAL);
	return ifStmt;
}

MatchClause *createMatchClause() {
	return safeCalloc(sizeof(MatchClause));
}

MatchStat *createMatchStat(Expression *expr) {
	MatchStat *match = safeMalloc(sizeof(*match));
	match->expr = expr;
	match->clauses = createVector(VECTOR_EXPONENTIAL);
	return match;
}

ForStat *createForStat() {
	return safeCalloc(sizeof(ForStat));
}

StructuredStatement *createStructuredStatement() {
	return safeCalloc(sizeof(StructuredStatement));
}

Statement *createStatement() {
	return safeCalloc(sizeof(Statement));
}

Type *createType() {
	return safeCalloc(sizeof(Type));
}

void cleanupAST(Vector *nodes) {
	for (int i = 0; i < nodes->size; i++) {
		Node *node = getVectorItem(nodes, i);

		switch (node->type) {
			case IDENTIFIER_LIST_NODE: destroyIdentifierList(node->data); break;
			case LITERAL_NODE: destroyLiteral(node->data); break;
			case UNARY_EXPR_NODE: destroyUnaryExpr(node->data); break;
			case MACRO_NODE: destroyMacro(node->data); break;
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
			case IMPL_NODE: destroyImpl(node->data); break;
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
			default: 
				printf("un-recognized node %d\n", node->type); 
				break;
		}
	}
}

void destroyUseMacro(UseMacro *use) {
	free(use);
}

void destroyArrayInitializer(ArrayInitializer *array) {
	for (int i = 0; i < array->values->size; i++) {
		destroyExpression(getVectorItem(array->values, i));
	}
	destroyVector(array->values);
	free(array);
}

void destroyLinkerFlagMacro(LinkerFlagMacro *linker) {
	free(linker);
}

void destroyIdentifierList(IdentifierList *list) {
	if (!list) return;
	destroyVector(list->values);
	free(list);
}

void destroyImpl(Impl *impl) {
	if (!impl) return;
	destroyVector(impl->funcs);
	free(impl);
}

void destroyLiteral(Literal *lit) {
	if (!lit) return;
	destroyCharLit(lit->charLit);
	destroyOtherLit(lit->otherLit);
	free(lit);
}

void destroyCharLit(CharLit *lit) {
	if (!lit) return;
	free(lit);
}

void destroyOtherLit(OtherLit *lit) {
	if (!lit) return;
	free(lit);
}

void destroyTupleType(TupleType *tuple) {
	for (int i = 0; i < tuple->types->size; i++) {
		destroyType(getVectorItem(tuple->types, i));
	}
	destroyVector(tuple->types);
	free(tuple);
}

void destroyOptionType(OptionType *type) {
	destroyType(type->type);
	free(type);
}

void destroyTupleExpr(TupleExpr *tuple) {
	for (int i = 0; i < tuple->values->size; i++) {
		destroyExpression(getVectorItem(tuple->values, i));
	}
	destroyVector(tuple->values);
	free(tuple);
}

void destroyUnaryExpr(UnaryExpr *expr) {
	if (!expr) return;
	destroyExpression(expr->lhand);
	free(expr);
}

void destroyEnumItem(EnumItem *item) {
	if (!item) return;
	destroyExpression(item->val);
	free(item);
}

void destroyEnumDecl(EnumDecl *decl) {
	if (!decl) return;
	for (int i = 0; i < decl->items->size; i++) {
		destroyEnumItem(getVectorItem(decl->items, i));
	}
	free(decl);
}

void destroyCall(Call *call) {
	if (!call) return;
	for (int i = 0; i < call->arguments->size; i++) {
		destroyExpression(getVectorItem(call->arguments, i));
	}
	destroyVector(call->arguments);
	destroyVector(call->callee);
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
	destroyType(arrayType->type);
	free(arrayType);
}

void destroyPointerType(PointerType *pointerType) {
	if (!pointerType) return;
	destroyType(pointerType->type);
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

void destroyFunctionSignature(FunctionSignature *func) {
	if (!func) return;
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
	free(assign);
}

void destroyUnstructuredStatement(UnstructuredStatement *stmt) {
	if (!stmt) return;
	switch (stmt->type) {
		case ASSIGNMENT_NODE: destroyAssignment(stmt->assignment); break;
		case DECLARATION_NODE: destroyDeclaration(stmt->decl); break;
		case INC_DEC_STAT_NODE: destroyIncDecStat(stmt->incDec); break;
		case LEAVE_STAT_NODE: destroyLeaveStat(stmt->leave); break;
		case FUNCTION_CALL_NODE: destroyCall(stmt->call); break;
		case IMPL_NODE: destroyImpl(stmt->impl); break;
		case EXPR_STAT_NODE: destroyExpression(stmt->expr); break;
		default:
			errorMessage("unstructured statement isn't being destroyed!?");
			break;
	}
	free(stmt);
}

void destroyMacro(Macro *macro) {
	if (!macro) return;
	switch (macro->type) {
		case USE_MACRO_NODE: destroyUseMacro(macro->use); break;
		case LINKER_FLAG_MACRO_NODE: destroyLinkerFlagMacro(macro->linker); break;
		default:
			errorMessage("macro not being destroyed?");
			break;
	}
	free(macro);
}

void destroyPointerFree(PointerFree *pntr) {
	free(pntr);
}

void destroyElseIfStat(ElseIfStat *stmt) {
	if (!stmt) return;
	destroyBlock(stmt->body);
	destroyExpression(stmt->condition);
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
	for (int i = 0; i < stmt->elseIfStmts->size; i++) {
		ElseIfStat *elseIfStat = getVectorItem(stmt->elseIfStmts, i);
		destroyElseIfStat(elseIfStat);
	}
	destroyVector(stmt->elseIfStmts);
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
	destroyExpression(stmt->index);
	destroyExpression(stmt->step);
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
