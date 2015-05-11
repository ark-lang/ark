#include "semantic.h"

VariableType deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call) {

}

VariableType deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit) {

}

VariableType deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr) {

}

VariableType deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr) {

}

VariableType deduceTypeFromExpression(SemanticAnalyzer *self, Expression *expr) {
	Vector *vec = createVector(VECTOR_EXPONENTIAL);

	switch (expr->exprType) {
		case BINARY_EXPR_NODE: deduceTypeFromBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: deduceTypeFromUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: deduceTypeFromFunctionCall(self, expr->call); break;
		case LITERAL_NODE: deduceTypeFromLiteral(self, expr->lit); break;
		default:
			errorMessage("Could not deduce type: %s", getNodeTypeName(expr->exprType));
			break;
	}
}
