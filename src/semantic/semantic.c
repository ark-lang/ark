#include "semantic.h"

// a lazy macro for failing when the error message is called.
#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

#define GLOBAL_SCOPE_INDEX 0

const char *VARIABLE_TYPE_NAMES[] = {
	"INTEGER_VAR_TYPE",
	"DOUBLE_VAR_TYPE",
	"STRING_VAR_TYPE",
	"STRUCTURE_VAR_TYPE",
	"CHAR_VAR_TYPE",
};

Scope *createScope() {
	Scope *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->structSymTable = hashmap_new();
	return self;
}

void destroyScope(Scope *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	hashmap_free(self->structSymTable);
	free(self);
}

SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles) {
	SemanticAnalyzer *self = safeMalloc(sizeof(*self));
	self->sourceFiles = sourceFiles;
	self->abstractSyntaxTree = NULL;
	self->currentSourceFile = NULL;
	self->currentNode = 0;
	self->scopes = createStack();
	self->failed = false;
	return self;
}

void analyzeBlock(SemanticAnalyzer *self, Block *block) {
	pushScope(self);

	for (int i = 0; i < block->stmtList->stmts->size; i++) {
		Statement *stmt = getVectorItem(block->stmtList->stmts, i);
		analyzeStatement(self, stmt);
	}

	popScope(self);
}

void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl) {
	pushFunctionDeclaration(self, decl);
	if (!decl->prototype) {
		analyzeBlock(self, decl->body);
	}
}

// wat
VariableType mergeTypes(VariableType a, VariableType b) {
	if (a == DOUBLE_VAR_TYPE || b == DOUBLE_VAR_TYPE) {
		return DOUBLE_VAR_TYPE;
	}
	else if (a == STRING_VAR_TYPE || b == STRING_VAR_TYPE) {
		return STRING_VAR_TYPE;
	}
	else {
		return INTEGER_VAR_TYPE;
	}
}

VariableType literalToType(Literal *literal) {
	switch (literal->type) {
		case LITERAL_WHOLE_NUMBER: return INTEGER_VAR_TYPE;
		case LITERAL_DECIMAL_NUMBER: return DOUBLE_VAR_TYPE;
		case LITERAL_HEX_NUMBER: return INTEGER_VAR_TYPE;
		case LITERAL_STRING: return STRING_VAR_TYPE;
		case LITERAL_CHAR: return CHAR_VAR_TYPE;
	}

	return false;
}

VariableType deduceType(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case LITERAL_NODE: return literalToType(expr->lit);
		case TYPE_NODE: {
			// type lit not supported atm TODO
			if (expr->type->type == TYPE_NAME_NODE) {
				VariableDecl *varDecl = checkVariableExists(self, expr->type->typeName->name);
				if (varDecl) {
					return deduceType(self, varDecl->expr);
				}
				else {
					semanticError("Could not deduce %s", expr->type->typeName->name);
				}
			}
			break;
		}
		case BINARY_EXPR_NODE: {
			// deduce type for lhand & rhand
			VariableType lhandType = deduceType(self, expr->binary->lhand);
			VariableType rhandType = deduceType(self, expr->binary->rhand);
			
			// merge them
			VariableType finalType = mergeTypes(lhandType, rhandType);
			return finalType;
		}
	}
	return false;
}

TypeName *createTypeDeduction(VariableType type) {
	switch (type) {
		case INTEGER_VAR_TYPE: return createTypeName("int");
		case DOUBLE_VAR_TYPE: return createTypeName("double");
		case STRING_VAR_TYPE: return createTypeName("str");
		case CHAR_VAR_TYPE: return createTypeName("char");
		default:
			printf("what to do for %s\n", VARIABLE_TYPE_NAMES[type]);
			break;
	}
	return false;
}

void analyzeVariableDeclaration(SemanticAnalyzer *self, VariableDecl *decl) {
	VariableDecl *mapDecl = checkVariableExists(self, decl->name);
	// doesnt exist
	if (!mapDecl) {
		pushVariableDeclaration(self, decl);
		
		if (decl->inferred) {
			VariableType type = deduceType(self, decl->expr);
			decl->type = createType();
			decl->type->type = TYPE_NAME_NODE;
			decl->type->typeName = createTypeDeduction(type);
		}
	}
	// does exist, oh shit
	else {
		semanticError("Redefinition of `%s`", decl->name);
	}
}

void analyzeAssignment(SemanticAnalyzer *self, Assignment *assign) {
	VariableDecl *mapDecl = checkVariableExists(self, assign->iden);
	// check assign thing exists
	if (mapDecl) {
		semanticError("`%s` undeclared", assign->iden);
	}
	// check mutability, etc.
	else {
		if (!mapDecl->mutable) {
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
	// probably some more shit we can do here, it'll do for now though
	analyzeExpression(self, expr->lhand);
	analyzeExpression(self, expr->rhand);
}

void analyzeUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {
	analyzeExpression(self, expr->lhand);
}

void analyzeExpression(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: analyzeBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: analyzeUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, expr->call); break;
	}
}

void analyzeUnstructuredStatement(SemanticAnalyzer *self, UnstructuredStatement *unstructured) {
	switch (unstructured->type) {
		case DECLARATION_NODE: analyzeDeclaration(self, unstructured->decl); break;
		case FUNCTION_CALL_NODE: analyzeFunctionCall(self, unstructured->call); break;
		case ASSIGNMENT_NODE: analyzeAssignment(self, unstructured->assignment); break;
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
	FunctionDecl *main = checkFunctionExists(self, MAIN_FUNC);
	if (!main) {
		semanticError("Undefined reference to `main`");
		self->failed = true;
	}
}

void startSemanticAnalysis(SemanticAnalyzer *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;
		
		// global scope
		pushScope(self);
		for (int j = 0; j < self->abstractSyntaxTree->size; j++) {
			Statement *stmt = getVectorItem(self->abstractSyntaxTree, j);
			analyzeStatement(self, stmt);
		}
		checkMainExists(self);
		popScope(self);
	}
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
	Scope *scope = getStackItem(self->scopes, self->scopes->stackPointer);
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
	hashmap_free(scope->varSymTable);
	hashmap_free(scope->structSymTable);
}

void destroySemanticAnalyzer(SemanticAnalyzer *self) {
	// clear up the remaining scope stuff
	for (int i = 0; i < self->scopes->stackPointer; i++) {
		popScope(self);
	}
	destroyStack(self->scopes);

	free(self);
}
