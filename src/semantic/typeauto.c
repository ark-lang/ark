#include "semantic.h"

VarType *createVarType(int type) {
	VarType *var = safeMalloc(sizeof(*var));
	var->type = type;
	return var;
}

void destroyVarType(VarType *type) {
	free(type);
}

VarType *deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call) {

}

VarType *deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit) {

}

VarType *deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {

}

VarType *deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {

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
			errorMessage("Types are not consistent, found: %d //TODO LOOKUP FOR THIS", i);
		}
	}
}
