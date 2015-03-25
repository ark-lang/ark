#include "ast.h"

IdentifierList *createIdentifierList() {
	IdentifierList *list = malloc(sizeof(*list));
	list->values = createVector();
	return list;
}

void destroyIdentifierList(IdentifierList *list) {
	destroyVector(list->values);
	free(list);
}

Expression *createExpression() {
	Expression *expr = malloc(sizeof(*expr));

	return expr;
}

void destroyExpression(Expression *expr) {
	free(expr);
}

TypeName *createTypeName() {
	TypeName *typeName = malloc(sizeof(*typeName));

	return typeName;
}

void destroyTypeName(TypeName *typeName) {
	free(typeName);
}

ArrayType *createArrayType() {
	ArrayType *arrayType = malloc(sizeof(*arrayType));

	return arrayType;
}

void destroyArrayType(ArrayType *arrayType) {
	destroyExpression(arrayType->length);
	free(arrayType);
}

PointerType *createPointerType() {
	PointerType *pointerType = malloc(sizeof(*pointerType));

	return pointerType;
}

void destroyPointerType(PointerType *pointerType) {
	free(pointerType);
}

FieldDecl *createFieldDecl() {
	FieldDecl *fieldDecl = malloc(sizeof(*fieldDecl));

	return fieldDecl;
}

void destroyFieldDecl(FieldDecl *fieldDecl) {
	free(fieldDecl);
}

FieldDeclList *createFieldDeclList() {
	FieldDeclList *fieldDeclList = malloc(sizeof(*fieldDeclList));

	return fieldDeclList;
}

void destroyFieldDeclList(FieldDeclList *fieldDeclList) {
	free(fieldDeclList);
}

StructType *createStructType() {
	StructType *structType = malloc(sizeof(*structType));

	return structType;
}

void destroyStructType(StructType *structType) {
	free(structType);
}

FunctionType *createFunctionType() {
	FunctionType *funcType = malloc(sizeof(*funcType));

	return funcType;
}

void destroyFunctionType(FunctionType *funcType) {
	free(funcType);
}

Type *createType() {
	Type *type = malloc(sizeof(*type));
	return type;
}

void destroyType(Type *type) {
	destroyArrayType(type->arrayType);
	destroyPointerType(type->pointerType);
	destroyStructType(type->structType);
	destroyTypeName(type->typeName);
}
