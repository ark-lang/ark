#include "semantic.h"

static const char *TYPE_NAME[] = {
	"int", "double", "str", "char"
};

VarType *createVarType(int type) {
	VarType *var = safeMalloc(sizeof(*var));
	var->type = type;
	var->isArray = false;
	var->arrayLen = 0;
	return var;
}

void destroyVarType(VarType *type) {
	free(type);
}

VarType *deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call) {
	char *funcName = getVectorItem(call->callee, 0);
	FunctionDecl *func = checkFunctionExists(self, funcName);
	if (func) {
		return deduceTypeFromType(self, func->signature->type);
	}
	return NULL;
}

VarType *deduceTypeFromTypeVal(SemanticAnalyzer *self, char *typeVal) {
	// int for now...
	for (int i = 0; i < ARR_LEN(TYPE_NAME); i++) {
		if (!strcmp(typeVal, TYPE_NAME[i])) {
			return createVarType(i);
		}
	}
	return NULL;
}

VarType *deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit) {
	switch (lit->type) {
		case CHAR_LITERAL_NODE: return createVarType(CHAR_VAR_TYPE);
		case STRING_LITERAL_NODE: return createVarType(STRING_VAR_TYPE);
		case INT_LITERAL_NODE: return createVarType(INTEGER_VAR_TYPE);
		case FLOAT_LITERAL_NODE: return createVarType(DOUBLE_VAR_TYPE);
	}
	return NULL;
}

VarType *deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {
	VariableType lhandType = deduceTypeFromExpression(self, expr->lhand)->type;
	VariableType rhandType = deduceTypeFromExpression(self, expr->rhand)->type;

	if (lhandType == rhandType) {
		return createVarType(lhandType);
	}
	errorMessage("Incompatible types %s and %s in binary expression <expression here?>", TYPE_NAME[lhandType], TYPE_NAME[rhandType]);
	return NULL;
}

VarType *deduceTypeFromType(SemanticAnalyzer *self, Type *type) {
	switch (type->type) {
		case TYPE_LIT_NODE:
			break;
		case TYPE_NAME_NODE: return deduceTypeFromTypeVal(self, type->typeName->name);
	}
	return NULL;
}

VarType *deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {
	printf("its a unary\n");
	return NULL;
}

VarType *deduceTypeFromTypeNode(SemanticAnalyzer *self, TypeName *type) {
	printf("type name is %s\n", type->name);
	VariableDecl *decl = checkVariableExists(self, type->name);
	if (decl) {
		return deduceTypeFromType(self, decl->type);
	}
	else {
		errorMessage("Type does not exist in the current/global scope: %s", type->name);
		return NULL;
	}
}

VarType *deduceTypeFromArrayInitializer(SemanticAnalyzer *self, ArrayInitializer *arr) {
	Vector *types = createVector(VECTOR_EXPONENTIAL);

	// store deduced types
	for (int i = 0; i < arr->values->size; i++) {
		Expression *expr = getVectorItem(arr->values, i);
		VariableType exprType = deduceTypeFromExpression(self, expr)->type;
		pushBackItem(types, &exprType);
	}

	// check for inconsistencies
	VariableType *firstType = getVectorItem(types, 0);
	for (int i = 0; i < types->size; i++) {
		VariableType *type = getVectorItem(types, i);
		if (type != firstType) {
			errorMessage("Inconsistent type in array initializer, found <TYPE>, expected <TYPE>\n");
			return NULL;
		}
	}
	destroyVector(types);

	VarType *result = createVarType(*firstType);
	result->isArray = true;
	result->arrayLen = arr->values->size;
	return result;
}

VarType *deduceTypeFromExpression(SemanticAnalyzer *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: return deduceTypeFromBinaryExpr(self, expr->binary);
		case UNARY_EXPR_NODE: return deduceTypeFromUnaryExpr(self, expr->unary);
		case FUNCTION_CALL_NODE: return deduceTypeFromFunctionCall(self, expr->call);
		case LITERAL_NODE: return deduceTypeFromLiteral(self, expr->lit);
		case TYPE_NODE: return deduceTypeFromTypeNode(self, expr->type->typeName);
		case ARRAY_INITIALIZER_NODE: return deduceTypeFromArrayInitializer(self, expr->arrayInitializer);
		default:
			errorMessage("Could not deduce type: %s", getNodeTypeName(expr->exprType));
			return NULL;
	}
}
