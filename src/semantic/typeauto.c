#include "semantic.h"

static const char *TYPE_NAME[] = {
	"int", "double", "str", "char"
};

VarType *createVarType(int type) {
	VarType *var = safeMalloc(sizeof(*var));
	var->type = type;
	return var;
}

void destroyVarType(VarType *type) {
	free(type);
}

VarType *deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call) {
	printf("func call\n");
}

VarType *deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit) {
	switch (lit->type) {
		case CHAR_LITERAL_NODE: return createVarType(CHAR_VAR_TYPE);
		case STRING_LITERAL_NODE: return createVarType(STRING_VAR_TYPE);
		case INT_LITERAL_NODE: return createVarType(INTEGER_VAR_TYPE);
		case FLOAT_LITERAL_NODE: return createVarType(DOUBLE_VAR_TYPE);
	}
}

VarType *deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {
	VariableType lhandType = deduceTypeFromExpression(self, expr->lhand);
	VariableType rhandType = deduceTypeFromExpression(self, expr->rhand);
	
	if (lhandType == rhandType) {
		return createVarType(lhandType);
	}
	errorMessage("incompatible types %s and %s in binary expression <expression here?>", TYPE_NAME[lhandType], TYPE_NAME[rhandType]);
	return NULL;
}

VarType *deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {
	printf("its a unary\n");
}

VariableType deduceTypeFromExpression(SemanticAnalyzer *self, Expression *expr) {
	Vector *vec = createVector(VECTOR_EXPONENTIAL);

	switch (expr->exprType) {
		case BINARY_EXPR_NODE: pushBackItem(vec, deduceTypeFromBinaryExpr(self, expr->binary)); break;
		case UNARY_EXPR_NODE: pushBackItem(vec, deduceTypeFromUnaryExpr(self, expr->unary)); break;
		case FUNCTION_CALL_NODE: pushBackItem(vec, deduceTypeFromFunctionCall(self, expr->call)); break;
		case LITERAL_NODE: pushBackItem(vec, deduceTypeFromLiteral(self, expr->lit)); break;
		default:
			errorMessage("Could not deduce type: %s", getNodeTypeName(expr->exprType));
			break;
	}

	//
	// Compare the types against the first value:
	// 5 + 2.3 + 9 -> wrong they aren't all integers (5 is an int)
	// maybe in this case coerce the type to a double? (numbers only)
	//

	// store the initialType
	int initialType = ((VarType*) getVectorItem(vec, 0))->type;
	for (int i = 1; i < vec->size; i++) {
		// compare types against the first type
		if (((VarType*) getVectorItem(vec, i))->type != initialType) {
			errorMessage("Types are not consistent, found: %d // <--- TODO LOOKUP FOR THIS??", i);
		}
	}

	return initialType;
}
