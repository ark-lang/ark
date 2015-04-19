#include "semantic.h"

// a lazy macro for failing when the error message is called.
#define semanticError(...) self->failed = true; \
						   errorMessage(__VA_ARGS__)

const char *VARIABLE_TYPE_NAMES[] = {
	"INTEGER_VAR_TYPE",
	"DOUBLE_VAR_TYPE",
	"STRING_VAR_TYPE",
	"STRUCTURE_VAR_TYPE",
	"CHAR_VAR_TYPE",
};

Scope *createScope() {
	Scope *self = safeMalloc(sizeof(*self));
	self->varSymTable = hashmap_new();
	self->structSymTable = hashmap_new();
	return self;
}

void destroyScope(Scope *self) {
	hashmap_free(self->varSymTable);
	hashmap_free(self->structSymTable);
	free(self);
}

SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles) {
	SemanticAnalyzer *self = safeMalloc(sizeof(*self));
	self->funcSymTable = hashmap_new();
	self->varSymTable = hashmap_new();
	self->structSymTable = hashmap_new();
	self->sourceFiles = sourceFiles;
	self->abstractSyntaxTree = NULL;
	self->currentSourceFile = NULL;
	self->currentNode = 0;
	self->failed = false;
	return self;
}

void analyzeBlock(SemanticAnalyzer *self, Block *block) {
	for (int i = 0; i < block->stmtList->stmts->size; i++) {
		Statement *stmt = getVectorItem(block->stmtList->stmts, i);
		analyzeStatement(self, stmt);
	}
}

void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl) {
	FunctionDecl *mapDecl = NULL;	
	if (hashmap_get(self->funcSymTable, decl->signature->name, (void**) &mapDecl) == MAP_MISSING) {
		hashmap_put(self->funcSymTable, decl->signature->name, decl);
		if (!decl->prototype) {
			analyzeBlock(self, decl->body);
		}
	}
	else {
		semanticError("Function with name `%s` already exists", decl->signature->name);
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
				VariableDecl *varDecl = checkGlobalVariableExists(self, expr->type->typeName->name);
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
	VariableDecl *mapDecl = NULL;	
	// doesnt exist
	if (hashmap_get(self->varSymTable, decl->name, (void**) &mapDecl) == MAP_MISSING) {
		hashmap_put(self->varSymTable, decl->name, decl);
		
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
	VariableDecl *mapDecl = NULL;
	// check assign thing exists
	if (hashmap_get(self->varSymTable, assign->iden, (void**) &mapDecl) == MAP_MISSING) {
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
		
		for (int j = 0; j < self->abstractSyntaxTree->size; j++) {
			Statement *stmt = getVectorItem(self->abstractSyntaxTree, j);
			analyzeStatement(self, stmt);
		}
	}

	checkMainExists(self);
}

StructDecl *checkGlobalStructureExists(SemanticAnalyzer *self, char *structName) {
	StructDecl *structure = NULL;
	if (hashmap_get(self->structSymTable, structName, (void**) &structure) == MAP_MISSING) {
		return false;
	}
	return structure;
}

VariableDecl *checkGlobalVariableExists(SemanticAnalyzer *self, char *varName) {
	VariableDecl *var = NULL;
	if (hashmap_get(self->funcSymTable, varName, (void**) &var) == MAP_MISSING) {
		return false;
	}
	return var;
}

FunctionDecl *checkFunctionExists(SemanticAnalyzer *self, char *funcName) {
	FunctionDecl *func = NULL;
	if (hashmap_get(self->funcSymTable, funcName, (void**) &func) == MAP_MISSING) {
		return false;
	}
	return func;
}

void destroySemanticAnalyzer(SemanticAnalyzer *self) {
	hashmap_free(self->funcSymTable);
	hashmap_free(self->varSymTable);
	free(self);
}
