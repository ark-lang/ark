#include "semantic.h"

// a lazy macro for failing when the error message is called.
#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

// global scope is always on the bottom of the stack
#define GLOBAL_SCOPE_INDEX 0

const char *VARIABLE_TYPE_NAMES[] = {
	"int",
	"double",
	"str",
	"char",
	"struct",
	"???"
};

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
	free(self);
}

VariableTypeHeap *createHeapType(VariableType type) {
	VariableTypeHeap *var = safeMalloc(sizeof(*var));
	var->type = type;
	return var;
}

SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles) {
	SemanticAnalyzer *self = safeMalloc(sizeof(*self));
	self->sourceFiles = sourceFiles;
	self->abstractSyntaxTree = NULL;
	self->currentSourceFile = NULL;
	self->currentNode = 0;
	self->scopes = createStack();
	self->failed = false;
	
	// data types
	self->dataTypes = hashmap_new();

	// this will do for now....

	// signed precision
	hashmap_put(self->dataTypes, "i8", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i16", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i32", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "i64", createHeapType(INTEGER_VAR_TYPE));

	// unsigned precision
	hashmap_put(self->dataTypes, "u8", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u16", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u32", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "u64", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "usize", createHeapType(INTEGER_VAR_TYPE));

	// precision floats
	hashmap_put(self->dataTypes, "f32", createHeapType(DOUBLE_VAR_TYPE));
	hashmap_put(self->dataTypes, "f64", createHeapType(DOUBLE_VAR_TYPE));

	// primitive types
	hashmap_put(self->dataTypes, "int", createHeapType(INTEGER_VAR_TYPE));
	hashmap_put(self->dataTypes, "char", createHeapType(CHAR_VAR_TYPE));
	hashmap_put(self->dataTypes, "str", createHeapType(STRING_VAR_TYPE));
	hashmap_put(self->dataTypes, "bool", createHeapType(INTEGER_VAR_TYPE));

	return self;
}

int getDataType(SemanticAnalyzer *self, char *type) {
	VariableTypeHeap *varType = NULL;
	if (hashmap_get(self->dataTypes, type, (void**) &varType) == MAP_OK) {
		return varType->type;
	}
	return -1;
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

// wat
VariableType mergeTypes(VariableType a, VariableType b) {
	if (a == DOUBLE_VAR_TYPE || b == DOUBLE_VAR_TYPE) {
		return DOUBLE_VAR_TYPE;
	}
	else if (a == INTEGER_VAR_TYPE && b == INTEGER_VAR_TYPE) {
		return INTEGER_VAR_TYPE;
	}
	else if (a == STRING_VAR_TYPE && b == STRING_VAR_TYPE) {
		return STRING_VAR_TYPE;
	}
	else if (a == CHAR_VAR_TYPE && b != CHAR_VAR_TYPE) {
		return INTEGER_VAR_TYPE;
	}
	else if (a == CHAR_VAR_TYPE && b == CHAR_VAR_TYPE) {
		return CHAR_VAR_TYPE;
	}
	return UNKNOWN_VAR_TYPE;
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

VariableType getTypeFromFunctionSignature(SemanticAnalyzer *self, Type *type) {
	if (type->type == TYPE_NAME_NODE) {
		VariableDecl *varDecl = checkVariableExists(self, type->typeName->name);
		int varType;

		if (varDecl) {
			return deduceType(self, varDecl->expr);
		}
		else if ((varType = getDataType(self, type->typeName->name)) != -1) {
			return varType;
		}
		else {
			semanticError("Could not deduce type based on function signature: `%s`", type->typeName->name);
		}
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
		case FUNCTION_CALL_NODE: {
			FunctionDecl *decl = checkFunctionExists(self, getVectorItem(expr->call->callee, 0));
			if (decl) {
				return getTypeFromFunctionSignature(self, decl->signature->type);
			}
			else {
				semanticError("Cannot deduce type for undefined function `%s`", getVectorItem(expr->call->callee, 0));
			}
			break;
		}
		case BINARY_EXPR_NODE: {
			// deduce type for lhand & rhand
			VariableType lhandType = deduceType(self, expr->binary->lhand);
			VariableType rhandType = deduceType(self, expr->binary->rhand);
			
			// merge them
			VariableType finalType = mergeTypes(lhandType, rhandType);
			if (finalType == UNKNOWN_VAR_TYPE) {
				semanticError("Incompatible types `%s` and `%s` in expression: ", VARIABLE_TYPE_NAMES[lhandType], VARIABLE_TYPE_NAMES[rhandType]);
			}
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
		default: break;
	}
	return false;
}

void analyzeVariableDeclaration(SemanticAnalyzer *self, VariableDecl *decl) {
	VariableDecl *mapDecl = checkVariableExists(self, decl->name);
	// doesnt exist
	if (!mapDecl) {
		if (decl->inferred) {
			VariableType type = deduceType(self, decl->expr);
			decl->type = createType();
			decl->type->type = TYPE_NAME_NODE;
			decl->type->typeName = createTypeDeduction(type);
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
