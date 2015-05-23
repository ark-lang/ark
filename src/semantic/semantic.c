#include "semantic.h"

// a lazy macro for failing when the error message is called.
#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

// global scope is always on the bottom of the stack
#define GLOBAL_SCOPE_INDEX 0

const char *reserved_keywords[] = {
	"func", "int", "float", "usize", "double",
	"str", "char", "rune", "i8", "u8", "i16",
	"u16", "i32", "u32", "i64", "u64", 
	"i128", "u128", "f32", "f64", "enum",
	"struct", "alloc", "free", "sizeof",
	"mut", "void", "impl", "if", "match",
	"for", "bool"
};

const char *valid_datatypes[] = {
	"int", "float", "usize", "double", "str", 
	"char", "rune", "i8", "u8", "i16", "u16", 
	"i32", "u32", "i64", "u64", "i128", "u128", 
	"f32", "f64" "void", "bool"
};

static bool isReservedKeyword(char *iden) {
	size_t keywords_size = ARR_LEN(reserved_keywords);
	for (size_t i = 0; i < keywords_size; i++) {
		if (!strcmp(iden, reserved_keywords[i])) {
			errorMessage("`%s` is a reserved keyword", iden);
			return true;
		}
	}
	return false;
}

static bool isValidDataType(char *dataType) {
	size_t datatypes_size = ARR_LEN(valid_datatypes);
	for (size_t i = 0; i < datatypes_size; i++) {
		if (!strcmp(dataType, valid_datatypes[i])) {
			return true;
		}
	}
	return false;
}

Scope *createScope() {
	Scope *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->paramSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->structSymTable = hashmap_new();
	self->implSymTable = hashmap_new();
	self->implFuncSymTable = hashmap_new();
	return self;
}

void destroyScope(Scope *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	hashmap_free(self->structSymTable);
	hashmap_free(self->paramSymTable);
	hashmap_free(self->implSymTable);
	hashmap_free(self->implFuncSymTable);
	free(self);
}

static char *getLoopIndex(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: {
			if (expr->binary->lhand) {
				char *name = getLoopIndex(self, expr->binary->lhand);
				isReservedKeyword(name);
				return name;
			}
			else {
				// TODO error here
			}
			break;
		}
		case TYPE_NODE: {
			switch (expr->type->type) {
				case TYPE_NAME_NODE: return expr->type->typeName->name;
			}
			break;
		}
	}
	errorMessage("Something went wrong in getLoopIndex");
	return NULL;
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
	hashmap_put(self->dataTypes, "i128", createVarType(INTEGER_VAR_TYPE));
	
	hashmap_put(self->dataTypes, "u8", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u16", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u32", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u64", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u128", createVarType(INTEGER_VAR_TYPE));

	hashmap_put(self->dataTypes, "f32", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "f64", createVarType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "f128", createVarType(INTEGER_VAR_TYPE));
	
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

bool isReturnStatement(Statement *stmt) {
	if (!stmt)
		return false;

	if (stmt->type != UNSTRUCTURED_STATEMENT_NODE)
		return false;

	UnstructuredStatement *unstrucStmt = stmt->unstructured;
	if (unstrucStmt->type != LEAVE_STAT_NODE)
		return false;

	return unstrucStmt->leave->type == RETURN_STAT_NODE;
}

void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl) {
	if (self->scopes->stackPointer != GLOBAL_SCOPE_INDEX) {
		semanticError("Function definition for `%s` is illegal", decl->signature->name);
	}

	isReservedKeyword(decl->signature->name);

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
		
		Vector *stmts = decl->body->stmtList->stmts;
		if (!(stmts->size > 0 && isReturnStatement(getVectorItem(stmts, stmts->size - 1)))) {
			pushBackItem(stmts, wrapExpressionInReturnStat(NULL));
		}
	}
}

void analyzeStructDeclaration(SemanticAnalyzer *self, StructDecl *decl) {
	isReservedKeyword(decl->name);

	pushStructDeclaration(self, decl);
}

Type *varTypeToType(VarType *type) {
	Type *result = createType();
	result->type = TYPE_NAME_NODE;

	switch (type->type) {
		case INTEGER_VAR_TYPE:
			result->typeName = createTypeName("int");
			break;
		case DOUBLE_VAR_TYPE:
			result->typeName = createTypeName("double");
			break;
		case STRING_VAR_TYPE:
			result->typeName = createTypeName("str");
			break;
		case CHAR_VAR_TYPE:
			result->typeName = createTypeName("i8");
			break;
		default: 
			errorMessage("Could not deduce type for <TODO> lol");
			return NULL; // TODO error
	}

	result->typeName->isArray = type->isArray;
	result->typeName->arrayLen = type->arrayLen;

	return result;
}

void analyzeVariableDeclaration(SemanticAnalyzer *self, VariableDecl *decl) {
	VariableDecl *mapDecl = checkVariableExists(self, decl->name);
	// doesn't exist
	if (!mapDecl) {
		isReservedKeyword(decl->name);

		if (decl->inferred) {
			VarType* type = deduceTypeFromExpression(self, decl->expr);
			decl->type = varTypeToType(type);
		}
		else if (decl->type->typeName) {
			// If we are initializing with a struct set the struct type
			StructDecl *structDecl = checkStructExists(self, decl->type->typeName->name);
			if (structDecl) {
				decl->structDecl = structDecl;
			}
			else {
				// not a structure
				if (!isValidDataType(decl->type->typeName->name)) {
					semanticError("`%s` is an invalid data type", decl->type->typeName->name);
				}
			}
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

void analyzeImplDeclaration(SemanticAnalyzer *self, ImplDecl *impl) {
	if (self->scopes->stackPointer != GLOBAL_SCOPE_INDEX) {
		semanticError("Impl declaration for `%s` is illegal", impl->name);
	}

	isReservedKeyword(impl->name);

	pushImplDeclaration(self, impl);

	// add all the impl functions to the impl function sym table
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	for (int i = 0; i < impl->funcs->size; ++i) {
		FunctionDecl *func = getVectorItem(impl->funcs, i);
		hashmap_put(scope->implFuncSymTable, func->signature->name, func);
	}
}

void analyzeDeclaration(SemanticAnalyzer *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: analyzeFunctionDeclaration(self, decl->funcDecl); break;
		case VARIABLE_DECL_NODE: analyzeVariableDeclaration(self, decl->varDecl); break;
		case IMPL_DECL_NODE: analyzeImplDeclaration(self, decl->implDecl); break;
		case STRUCT_DECL_NODE: analyzeStructDeclaration(self, decl->structDecl); break;
	}
}

void analyzeFunctionCall(SemanticAnalyzer *self, Call *call) {
	char *callee = getVectorItem(call->callee, 0);

	isReservedKeyword(callee);

	// typecasting
	Type *type = checkDataTypeExists(self, callee);
	if (type) {
		if (call->arguments->size > 1) {
			semanticError("Typecast can only take 1 argument");
		}
		else if (call->arguments->size == 0) {
			semanticError("Typecast needs at least 1 argument");
		}
		else {
			// TODO use actual type instead of char if types in semantic are better
			call->typeCast = callee;
		}
	}
	else {
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

			// Analyze every expression in a function call
			for (int i = 0; i < call->arguments->size; ++i) {
				analyzeExpression(self, getVectorItem(call->arguments, i));
			}
		}
	}
}

void analyzeBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {
	/*
	 * If we have an expression like X.Y() we need to
	 * set the actual struct type of X so the code generator can use it to
	 * emit the correct function name.
	 */
	if (expr->lhand->exprType == TYPE_NODE && expr->rhand->exprType == FUNCTION_CALL_NODE) {
		VariableDecl *decl = checkVariableExists(self, expr->lhand->type->typeName->name);
		if (decl) {
			expr->structType = decl->type;
		}
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
		case TYPE_NAME_NODE: 
			printf("penis: %s\n", type->typeName->name);
			break;
	}
}

void analyzeLiteralNode(SemanticAnalyzer *self, Literal *lit) {
	// TODO
	switch (lit->type) {
		case CHAR_LITERAL_NODE: break;
		case INT_LITERAL_NODE: break;
	}
}

void analyzeExpression(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: analyzeBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: analyzeUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, expr->call); break;
		case TYPE_NODE: analyzeTypeNode(self, expr->type); break;
		case LITERAL_NODE: analyzeLiteralNode(self, expr->lit); break;
		case SIZEOF_NODE: analyzeExpression(self, expr->sizeOf->expr); break;
		case ARRAY_INDEX_NODE: analyzeExpression(self, expr->arrayIndex->index); break;
		default:
			errorMessage("Unknown node in expression: %s (%d)", getNodeTypeName(expr->exprType), expr->exprType);
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
		case FREE_STAT_NODE: break; //TODO
		default:
			errorMessage("Unknown node in unstructured statement: %s", getNodeTypeName(unstructured->type));
			break;
	}
}

void analyzeForStat(SemanticAnalyzer *self, ForStat *stmt) {
	switch (stmt->forType) {
		case INDEX_FOR_LOOP: {
			if (stmt->index->binary->lhand) {
				char *iden = getLoopIndex(self, stmt->index->binary->lhand);
				VariableDecl *decl = checkVariableExists(self, iden);
				if (!decl) {
					errorMessage("For loop index `%s` does not exist in local or global scope", iden);
				}
				else if (!decl->mutable) {
					errorMessage("For loop index `%s` is not mutable", iden);
				}
			}
			else {
				errorMessage("Couldn't get for loop index");
			}
			break;
		}
		default: break;
	}
}

void analyzeStructuredStatement(SemanticAnalyzer *self, StructuredStatement *structured) {
	switch (structured->type) {
		case IF_STAT_NODE: analyzeIfStatement(self, structured->ifStmt); break;
		case FOR_STAT_NODE: analyzeForStat(self, structured->forStmt); break;
	}
	// TODO eventually display an errorMessage here if all the structured stmts are implemented
}

void analyzeIfStatement(SemanticAnalyzer *self, IfStat *ifStmt) {
	analyzeExpression(self, ifStmt->expr);
	analyzeBlock(self, ifStmt->body);

	if (ifStmt->elseIfStmts != NULL) {
		for (int i = 0; i < ifStmt->elseIfStmts->size; ++i) {
			analyzeElseIfStatement(self, getVectorItem(ifStmt->elseIfStmts, i));
		}
	}

	if (ifStmt->elseStmt != NULL) {
		analyzeElseStatement(self, ifStmt->elseStmt);
	}
}

void analyzeElseIfStatement(SemanticAnalyzer *self, ElseIfStat *elseIfStmt) {
	analyzeExpression(self, elseIfStmt->expr);
	analyzeBlock(self, elseIfStmt->body);
}

void analyzeElseStatement(SemanticAnalyzer *self, ElseStat *elseStmt) {
	analyzeBlock(self, elseStmt->body);
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
	if (!mainDecl && !IGNORE_MAIN) {
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

StructDecl *checkStructExists(SemanticAnalyzer *self, char *structName) {
	StructDecl *decl = checkLocalStructExists(self, structName);
	if (decl) {
		return decl;
	}

	decl = checkGlobalStructExists(self, structName);
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

StructDecl *checkLocalStructExists(SemanticAnalyzer *self, char *structName) {
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

StructDecl *checkGlobalStructExists(SemanticAnalyzer *self, char *structName) {
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

ImplDecl *checkGlobalImplExists(SemanticAnalyzer *self, char *implName) {
	ImplDecl *impl = NULL;
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (hashmap_get(scope->implSymTable, implName, (void**) &impl) == MAP_OK) {
		return impl;
	}
	return false;
}

FunctionDecl *checkFunctionExists(SemanticAnalyzer *self, char *funcName) {
	FunctionDecl *func = NULL;
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (hashmap_get(scope->funcSymTable, funcName, (void**) &func) == MAP_OK) {
		return func;
	}
	if (hashmap_get(scope->implFuncSymTable, funcName, (void**) &func) == MAP_OK) {
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

Type *checkDataTypeExists(SemanticAnalyzer *self, char *name) {
	Type *type = NULL;
	if (hashmap_get(self->dataTypes, name, (void**) &type) == MAP_OK) {
		return type;
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

void pushStructDeclaration(SemanticAnalyzer *self, StructDecl *structure) {
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

void pushImplDeclaration(SemanticAnalyzer *self, ImplDecl *impl) {
	Scope *scope = getStackItem(self->scopes, GLOBAL_SCOPE_INDEX);
	if (checkGlobalImplExists(self, impl->name)) {
		semanticError("Impl with the name `%s` has already been defined", impl->name);
		return;
	}
	hashmap_put(scope->implSymTable, impl->name, impl);
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
	hashmap_free(scope->implFuncSymTable);
	hashmap_free(scope->implSymTable);
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
