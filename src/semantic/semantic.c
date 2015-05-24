#include "semantic.h"

// a lazy macro for failing when the error message is called.
#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

// global scope is always on the bottom of the stack
#define GLOBAL_SCOPE_INDEX 0

Scope *createScope() {
	Scope *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->paramSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->structSymTable = hashmap_new();
	return self;
}

void destroyScope(Scope *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	hashmap_free(self->structSymTable);
	hashmap_free(self->paramSymTable);
	free(self);
}

SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles) {
	SemanticAnalyzer *self = safeMalloc(sizeof(*self));
	self->sourceFiles = sourceFiles;
	self->abstractSyntaxTree = NULL;
	self->currentSourceFile = NULL;
	self->currentNode = 0;
	self->scopes = createStack();
	self->dataTypes = hashmap_new();
	self->failed = false;

	hashmap_put(self->dataTypes, "i8", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i16", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i32", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i64", createVarType(INTEGER_VAR_TYPE));

	hashmap_put(self->dataTypes, "u8", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u16", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u32", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u64", createVarType(INTEGER_VAR_TYPE));

	hashmap_put(self->dataTypes, "f32", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "f64", createVarType(INTEGER_VAR_TYPE));

	hashmap_put(self->dataTypes, "int", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "char", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "str", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "bool", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "float", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "double", createVarType(INTEGER_VAR_TYPE));

	return self;
}

void analyzeBlock(SemanticAnalyzer *self, Block *block) {
	for (int i = 0; i < block->stmtList->stmts->size; i++) {
		Statement *stmt = getVectorItem(block->stmtList->stmts, i);
		analyzeStatement(self, stmt);
	}
}

void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl) {
	if (self->scopes->stackPointer != GLOBAL_SCOPE_INDEX) {
		semanticError("Function definition for `%s` is illegal", decl->signature->name);
	}

	pushFunctionDeclaration(self, decl);
	if (!decl->prototype) {
		pushScope(self);
	}
	for (int i = 0; i < decl->signature->parameters->paramList->size; i++) {
		ParameterSection *param = getVectorItem(decl->signature->parameters->paramList, i);
		pushParameterSection(self, param);
	}
	if (!decl->prototype) {
		analyzeBlock(self, decl->body);
		popScope(self);
	}
}

Type *varTypeToType(VariableType type) {
	Type *result = createType();
	result->type = TYPE_NAME_NODE;

	switch (type) {
		case INTEGER_VAR_TYPE:
			result->typeName = createTypeName("int");
			break;
		case DOUBLE_VAR_TYPE:
			result->typeName = createTypeName("double");
			break;
		case STRING_VAR_TYPE:
			result->typeName = createTypeName("str");
			break;
		// TODO eventually runes or whatever when we do utf8???
		case CHAR_VAR_TYPE:
			result->typeName = createTypeName("i8");
			break;
		default: return NULL; // TODO error
	}

	return result;
}

void analyzeVariableDeclaration(SemanticAnalyzer *self, VariableDecl *decl) {
	VariableDecl *mapDecl = checkVariableExists(self, decl->name);
	// doesnt exist
	if (!mapDecl) {
		if (decl->inferred) {
			VariableType type = deduceTypeFromExpression(self, decl->expr);
			decl->type = varTypeToType(type);
		}
		pushVariableDeclaration(self, decl);
	}
	// does exist, oh shit
	else {
		semanticError("Redefinition of `%s`", decl->name);
	}
}

void analyzeAssignment(SemanticAnalyzer *self, Assignment *assign) {
	VariableDecl *mapDecl = checkVariableExists(self, assign->iden);
	ParameterSection *paramDecl = checkLocalParameterExists(self, assign->iden);

	// check assign thing exists
	if (!mapDecl && !paramDecl) {
		semanticError("`%s` undeclared", assign->iden);
	}
	// check mutability, etc.
	else {
		if ((mapDecl && !mapDecl->mutable) || (paramDecl && !paramDecl->mutable)) {
			semanticError("Assignment of read-only variable `%s`", assign->iden);
		}
	}
}

void analyzeDeclaration(SemanticAnalyzer *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: analyzeFunctionDeclaration(self, decl->funcDecl); break;
		case VARIABLE_DECL_NODE: analyzeVariableDeclaration(self, decl->varDecl); break;
	}
}

void analyzeFunctionCall(SemanticAnalyzer *self, Call *call) {
	char *callee = getVectorItem(call->callee, 0);

	FunctionDecl *decl = checkFunctionExists(self, callee);
	if (!decl) {
		semanticError("Attempting to call undefined function `%s`", callee);
	}
	// it exists, check arguments match in length
	else {
		int argsNeeded = decl->signature->parameters->paramList->size;
		int argsGot = call->arguments->size;
		char *callee = getVectorItem(call->callee, 0); // FIXME

		// only do this on non-variadic functions, otherwise
		// it will fuck you over since variadic is variable amount of args
		if (!decl->signature->parameters->variadic) {
			if (argsGot > argsNeeded) {
				semanticError("Too many arguments to function `%s`", callee);
			}
			else if (argsGot < argsNeeded) {
				semanticError("Too few arguments to function `%s`", callee);
			}
		}
	}
}

void analyzeBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {
	// assignment (probably)
	if (!strcmp(expr->binaryOp, "=")) {
		printf("its an assignment!\n");
		// TODO lol
	}

	analyzeExpression(self, expr->lhand);
	analyzeExpression(self, expr->rhand);
}

void analyzeUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {
	analyzeExpression(self, expr->lhand);
}

void analyzeTypeNode(SemanticAnalyzer *self, Type *type) {
 	// TODO
	switch (type->type) {
		case TYPE_LIT_NODE: break;
		case TYPE_NAME_NODE: break;
	}
}

void analyzeLiteralNode(SemanticAnalyzer *self, Literal *lit) {
	switch (lit->type) {
		case CHAR_LITERAL_NODE:
			printf("char literal value: %d\n", lit->charLit->value);
			break;
		default:
			printf("other literal value: %s\n", lit->stringLit->value);
	}
}

void analyzeExpression(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: analyzeBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: analyzeUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, expr->call); break;
		case TYPE_NODE: analyzeTypeNode(self, expr->type); break;
		case LITERAL_NODE: analyzeLiteralNode(self, expr->lit); break;
		default:
			errorMessage("Unkown node in expression: %s (%d)", getNodeTypeName(expr->exprType), expr->exprType);
			break;
	}
}

void analyzeUnstructuredStatement(SemanticAnalyzer *self, UnstructuredStatement *unstructured) {
	switch (unstructured->type) {
		case DECLARATION_NODE: analyzeDeclaration(self, unstructured->decl); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, unstructured->call); break;
		case ASSIGNMENT_NODE: analyzeAssignment(self, unstructured->assignment); break;
		case EXPR_STAT_NODE: analyzeExpression(self, unstructured->expr); break;
		case LEAVE_STAT_NODE: break; // TODO
		default:
			errorMessage("Unkown node in expression: %s", getNodeTypeName(unstructured->type));
			break;
	}
}

void analyzeStructuredStatement(SemanticAnalyzer *self, StructuredStatement *structured) {
	ALLOY_UNUSED_OBJ(self);
	ALLOY_UNUSED_OBJ(structured);
	// TODO:
}

void analyzeStatement(SemanticAnalyzer *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE:
			analyzeUnstructuredStatement(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE:
			analyzeStructuredStatement(self, stmt->structured);
			break;
	}
}

void checkMainExists(SemanticAnalyzer *self) {
	FunctionDecl *mainDecl = checkFunctionExists(self, MAIN_FUNC);
	if (!mainDecl) {
		semanticError("Undefined reference to `main`");
	}
}

void startSemanticAnalysis(SemanticAnalyzer *self) {
	// global scope
	pushScope(self);
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		for (int j = 0; j < self->abstractSyntaxTree->size; j++) {
			Statement *stmt = getVectorItem(self->abstractSyntaxTree, j);
			analyzeStatement(self, stmt);
		}
	}
	checkMainExists(self);
	popScope(self);
}

StructDecl *checkStructureExists(SemanticAnalyzer *self, char *structName) {
	StructDecl *decl = checkLocalStructureExists(self, structName);
	if (decl) {
		return decl;
	}

	decl = checkGlobalStructureExists(self, structName);
	if (decl) {
		return decl;
	}

	return false;
}

VariableDecl *checkVariableExists(SemanticAnalyzer *self, char *varName) {
	VariableDecl *decl = checkLocalVariableExists(self, varName);
	if (decl) {
		return decl;
	}

	decl = checkGlobalVariableExists(self, varName);
	if (decl) {
		return decl;
	}

	return false;
}

StructDecl *checkLocalStructureExists(SemanticAnalyzer *self, char *structName) {
	StructDecl *structure = NULL;
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (hashmap_get(scope->structSymTable, structName, (void**) &structure) == MAP_OK) {
		return structure;
	}
	return false;
}

VariableDecl *checkLocalVariableExists(SemanticAnalyzer *self, char *varName) {
	VariableDecl *var = NULL;
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (hashmap_get(scope->varSymTable, varName, (void**) &var) == MAP_OK) {
		return var;
	}
	return false;
}

StructDecl *checkGlobalStructureExists(SemanticAnalyzer *self, char *structName) {
	StructDecl *structure = NULL;
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (hashmap_get(scope->structSymTable, structName, (void**) &structure) == MAP_OK) {
		return structure;
	}
	return false;
}

VariableDecl *checkGlobalVariableExists(SemanticAnalyzer *self, char *varName) {
	VariableDecl *var = NULL;
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (hashmap_get(scope->varSymTable, varName, (void**) &var) == MAP_OK) {
		return var;
	}
	return false;
}

FunctionDecl *checkFunctionExists(SemanticAnalyzer *self, char *funcName) {
	FunctionDecl *func = NULL;
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (hashmap_get(scope->funcSymTable, funcName, (void**) &func) == MAP_OK) {
		return func;
	}
	return false;
}

ParameterSection *checkLocalParameterExists(SemanticAnalyzer *self, char *paramName) {
	ParameterSection *param = NULL;
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (hashmap_get(scope->paramSymTable, paramName, (void**) &param) == MAP_OK) {
		return param;
	}
	return false;
}

void pushParameterSection(SemanticAnalyzer *self, ParameterSection *param) {
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (self->scopes->stackPointer == GLOBAL_SCOPE_INDEX) return;
	if (!checkLocalParameterExists(self, param->name)) {
		hashmap_put(scope->paramSymTable, param->name, param);
	}
}

void pushVariableDeclaration(SemanticAnalyzer *self, VariableDecl *var) {
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (checkLocalVariableExists(self, var->name)) {
		semanticError("Variable with the name `%s` already exists locally", var->name);
		return;
	}
	hashmap_put(scope->varSymTable, var->name, var);
}

void pushStructureDeclaration(SemanticAnalyzer *self, StructDecl *structure) {
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
	if (checkLocalVariableExists(self, structure->name)) {
		semanticError("Structure with the name `%s` already exists locally", structure->name);
		return;
	}
	hashmap_put(scope->structSymTable, structure->name, structure);
}

void pushFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *func) {
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (checkLocalVariableExists(self, func->signature->name)) {
		semanticError("Function with the name `%s` has already been defined", func->signature->name);
		return;
	}
	hashmap_put(scope->funcSymTable, func->signature->name, func);
}

void pushScope(SemanticAnalyzer *self) {
	pushToStack(self->scopes, createScope());
}

void popScope(SemanticAnalyzer *self) {
	Scope *scope = popStack(self->scopes);
	hashmap_free(scope->funcSymTable);
	hashmap_free(scope->paramSymTable);
	hashmap_free(scope->varSymTable);
	hashmap_free(scope->structSymTable);
	free(scope);
}

void destroySemanticAnalyzer(SemanticAnalyzer *self) {
	// clear up the remaining scope stuff
	for (int i = 0; i < self->scopes->stackPointer; i++) {
		popScope(self);
	}
	hashmap_free(self->dataTypes);
	destroyStack(self->scopes);

	free(self);
}
